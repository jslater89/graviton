package data

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Batch struct {
	ID              bson.ObjectId    `bson:"_id,omitempty"`
	RecipeName      string           `bson:"recipe"`
	UniqueID        string           `bson:"stringId"`
	HydrometerID    bson.ObjectId    `bson:"hydrometer"`
	GravityReadings []GravityReading `bson:"readings"`
	StartDate       time.Time        `bson:"startDate"`
	LastUpdate      time.Time        `bson:"lastUpdate"`

	Active bool `bson:"active"`
}

type GravityReading struct {
	ID             bson.ObjectId `json:"id" bson:"_id"`
	Date           time.Time     `json:"date" bson:"date"`
	Gravity        float64       `json:"gravity" bson:"gravity"`
	Temperature    float64       `json:"temperature" bson:"temperature"`
	BatteryVoltage float64       `json:"battery" bson:"battery"`
	Hidden         bool          `json:"hidden" bson:"hidden"`
}

type Hydrometer struct {
	ID             bson.ObjectId `bson:"_id,omitempty"`
	Name           string        `bson:"name"`
	Description    string        `bson:"description"`
	CurrentBatchID bson.ObjectId `bson:"batch"`
}
