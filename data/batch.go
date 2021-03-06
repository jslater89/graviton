package data

import (
	"errors"
	"time"

	"gopkg.in/mgo.v2"

	"github.com/jslater89/graviton"
	"go.uber.org/zap"

	"gopkg.in/mgo.v2/bson"
)

func (b *Batch) Save() error {
	if err := b.verify(); err != nil {
		return err
	}

	if b.ID == "" {
		b.ID = bson.NewObjectId()
	}

	_, err := db.batchCollection.UpsertId(b.ID, *b)

	if err != nil {
		return err
	}
	return nil
}

func SingleBatch(query bson.M) (*Batch, error) {
	batches, err := QueryBatches(query)

	if err != nil {
		return nil, err
	}

	// Allow the user to decide in this case
	if len(batches) > 1 {
		return batches[0], errors.New("single batch query returned multiple batches")
	}

	if len(batches) < 1 {
		return nil, mgo.ErrNotFound
	}

	return batches[0], nil
}

func QueryBatches(query bson.M) ([]*Batch, error) {
	batches := []*Batch{}
	err := db.batchCollection.Find(query).All(&batches)
	return batches, err
}

func AddBatch(b *Batch) (*Batch, error) {
	newBatch := &Batch{
		RecipeName:      b.RecipeName,
		StartDate:       b.StartDate,
		UniqueID:        b.UniqueID,
		Active:          b.Active,
		Archived:        b.Archived,
		GravityReadings: []GravityReading{},
	}

	err := newBatch.Save()

	if err != nil {
		return nil, err
	}

	err = newBatch.SetHydrometerID(b.HydrometerID)

	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	}

	return newBatch, nil
}

func (b *Batch) SetHydrometerID(hID bson.ObjectId) error {
	if hID == "" {
		hID = graviton.EmptyID()
	}

	hydrometer, err := SingleHydrometer(bson.M{"_id": hID})

	if err != nil && err != mgo.ErrNotFound {
		return err
	} else if err != nil && err == mgo.ErrNotFound {
		return b.SetHydrometer(&Hydrometer{ID: graviton.EmptyID()})
	}

	return b.SetHydrometer(hydrometer)
}

func (b *Batch) SetHydrometer(h *Hydrometer) error {
	// TODO: test case for this block: setting a hydrometer on a batch
	// should unset the batch's original hydrometer's batch
	if b.HydrometerID != "" && b.HydrometerID != graviton.EmptyID() {
		hydrometer, err := SingleHydrometer(bson.M{"_id": b.HydrometerID})

		if err != nil {
			return err
		}

		hydrometer.CurrentBatchID = graviton.EmptyID()
		err = hydrometer.Save()
		if err != nil {
			return err
		}
	}

	b.HydrometerID = h.ID
	h.CurrentBatchID = b.ID

	err := b.Save()
	if err != nil {
		return err
	}

	if h.ID != "" && h.ID != graviton.EmptyID() {
		err = h.Save()
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Batch) AddReading(r GravityReading) error {
	r.Hidden = false

	if r.ID == "" {
		r.ID = bson.NewObjectId()
	}

	// Insert at the correct time
	insertBefore := 0
	for i, reading := range b.GravityReadings {
		insertBefore = i
		if r.Date.Before(reading.Date) {
			break
		}

		if i == len(b.GravityReadings) - 1 {
			insertBefore = -1;
		}
	}

	var readings []GravityReading
	if insertBefore > 0 {
		readings = append(b.GravityReadings[:insertBefore], append([]GravityReading{r}, b.GravityReadings[insertBefore:]...)...)
	} else {
		readings = append(b.GravityReadings, r)
	}

	b.GravityReadings = readings
	b.LastUpdate = time.Now()
	return b.Save()
}

func (b *Batch) HideReadingID(id bson.ObjectId) error {
	for i, reading := range b.GravityReadings {
		if reading.ID == id {
			reading.Hidden = true

			readings := append(b.GravityReadings[:i], reading)
			readings = append(readings, b.GravityReadings[i+1:]...)
			b.GravityReadings = readings
			return b.Save()
		}
	}
	return errors.New("reading not found")
}

func (b *Batch) FinishBatch() error {
	b.Active = false
	b.LastUpdate = time.Now()

	hydrometers, err := QueryHydrometers(bson.M{"_id": b.HydrometerID})

	if err != nil {
		return err
	}

	h := hydrometers[0]
	h.CurrentBatchID = graviton.EmptyID()

	err = b.Save()
	if err != nil {
		return err
	}

	return h.Save()
}

func (b *Batch) ArchiveBatch() error {
	if b.Active {
		err := b.FinishBatch()
		if err != nil {
			return err
		}
	}

	b.Archived = true
	return b.Save()
}

func (b *Batch) CalculateGravityDelta() float64 {
	return 1.0
}

func (b *Batch) verify() error {
	if b.HydrometerID == "" {
		b.HydrometerID = graviton.EmptyID()
	} else if b.HydrometerID != graviton.EmptyID() {
		query := bson.M{
			"hydrometer": b.HydrometerID,
			"active":     true,
		}

		batches, err := QueryBatches(query)

		if err != nil {
			graviton.Logger.Error("Failed to query batches", zap.Error(err))
			return errors.New("verify query failed")
		}
		if len(batches) > 1 {
			return errors.New("hydrometer assigned to too many batches")
		}

		if len(batches) == 1 && batches[0].ID != b.ID {
			return errors.New("hydrometer already assigned to a different batch")
		}
	}

	return nil
}
