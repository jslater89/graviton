package data

import "gopkg.in/mgo.v2/bson"

func (b *Batch) Save() error {
	return nil
}

func QueryBatches(query bson.M) error {
	return nil
}

func (b *Batch) SetHydrometer(h *Hydrometer) error {
	// if an active batch has the hydrometer
	query := bson.M{
		"active":     true,
		"hydrometer": h.ID,
	}

	db.hydrometerCollection.Find(query).Count()

	b.HydrometerID = h.ID
	h.CurrentBatchID = b.ID

	if err := b.verify(); err != nil {
		return err
	}

	if err := h.verify(); err != nil {
		return err
	}

	b.Save()
	h.Save()

	return nil
}

func (b *Batch) verify() error {
	return nil
}
