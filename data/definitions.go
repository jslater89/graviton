package data

import (
	"time"

	"gopkg.in/mgo.v2/bson"
)

type Batch struct {
	ID              bson.ObjectId    `json:"_id" bson:"_id"`
	RecipeName      string           `json:"recipe" bson:"recipe"`
	UniqueID        string           `json:"stringId" bson:"stringId"`
	HydrometerID    bson.ObjectId    `json:"hydrometer" bson:"hydrometer"`
	InitialReading  GravityReading   `json:"initialReading,omitempty" bson:"initialReading,omitempty"`
	GravityReadings []GravityReading `json:"readings" bson:"readings"`
	StartDate       time.Time        `json:"startDate" bson:"startDate"`
	LastUpdate      time.Time        `json:"lastUpdate" bson:"lastUpdate"`

	Active bool `json:"active" bson:"active"`
}

type GravityReading struct {
	Date        time.Time `json:"date" bson:"date"`
	Gravity     float64   `json:"gravity" bson:"gravity"`
	Temperature float64   `json:"temperature" bson:"temperature"`
}

type Hydrometer struct {
	ID             bson.ObjectId `json:"_id" bson:"_id"`
	Name           string        `bson:"name" json:"name"`
	Description    string        `bson:"description" json:"description"`
	CurrentBatchID bson.ObjectId `bson:"batch" json:"batch"`
}

func (b *Batch) CalculateGravity() float64 {
	return 1.0
}

func (b *Batch) AddReading(r GravityReading) {

}
