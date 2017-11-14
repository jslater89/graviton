package data

import (
	"errors"

	"github.com/jslater89/graviton"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

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

func QueryHydrometers(query bson.M) ([]*Hydrometer, error) {
	hydrometers := []*Hydrometer{}
	err := db.hydrometerCollection.Find(query).All(&hydrometers)
	return hydrometers, err
}

func SingleHydrometer(query bson.M) (*Hydrometer, error) {
	hydrometers, err := QueryHydrometers(query)

	if err != nil {
		return nil, err
	}

	// Allow the user to decide in this case
	if len(hydrometers) > 1 {
		return hydrometers[0], errors.New("single hydrometer query returned multiple hydrometers")
	}

	if len(hydrometers) < 1 {
		return nil, mgo.ErrNotFound
	}

	return hydrometers[0], nil
}

func (h *Hydrometer) verify() error {
	return nil
}
