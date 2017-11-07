package data

import (
	"errors"
	"time"

	"github.com/jslater89/graviton"

	"gopkg.in/mgo.v2/bson"
)

func (b *Batch) Save() error {
	if err := b.verify(); err != nil {
		return err
	}
	_, err := db.batchCollection.UpsertId(b.ID, *b)

	if err != nil {
		return err
	}
	return nil
}

func QueryBatches(query bson.M) ([]Batch, error) {
	batches := []Batch{}
	err := db.batchCollection.Find(query).All(&batches)
	return batches, err
}

func (b *Batch) SetHydrometer(h *Hydrometer) error {
	b.HydrometerID = h.ID
	h.CurrentBatchID = b.ID

	err := b.Save()
	if err != nil {
		return err
	}

	if h.ID != graviton.EmptyID() {
		err = h.Save()
		if err != nil {
			return err
		}
	}

	return nil
}

func (b *Batch) AddReading(r GravityReading) {
	b.GravityReadings = append(b.GravityReadings, r)
	b.LastUpdate = time.Now()
	b.Save()
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

	b.Save()
	h.Save()

	return nil
}

func (b *Batch) CalculateGravity() float64 {
	return 1.0
}

func (b *Batch) verify() error {
	query := bson.M{
		"hydrometer": b.HydrometerID,
		"active":     true,
	}

	batches, err := QueryBatches(query)

	if err != nil {
		return errors.New("verify query failed")
	}

	if len(batches) > 1 {
		return errors.New("hydrometer assigned to too many batches")
	}

	if len(batches) == 1 && batches[0].ID != b.ID {
		return errors.New("hydrometer already assigned to a different batch")
	}

	return nil
}
