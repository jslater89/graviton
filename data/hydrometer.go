package data

import "gopkg.in/mgo.v2/bson"
import "github.com/jslater89/graviton"

func (h *Hydrometer) Save() error {
	if err := h.verify(); err != nil {
		return err
	}

	if h.CurrentBatchID == "" {
		h.CurrentBatchID = graviton.EmptyID()
	}
	_, err := db.hydrometerCollection.UpsertId(h.ID, *h)

	if err != nil {
		return err
	}
	return nil
}

func QueryHydrometers(query bson.M) ([]Hydrometer, error) {
	hydrometers := []Hydrometer{}
	err := db.hydrometerCollection.Find(query).All(&hydrometers)
	return hydrometers, err
}

func (h *Hydrometer) verify() error {
	return nil
}
