package api

import (
	"time"

	"github.com/jslater89/graviton"
	"github.com/jslater89/graviton/data"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

type Batch struct {
	ID              bson.ObjectId          `json:"id"`
	RecipeName      string                 `json:"recipe"`
	UniqueID        string                 `json:"stringId"`
	Hydrometer      Hydrometer             `json:"hydrometer"`
	GravityReadings *[]data.GravityReading `json:"readings"`
	LatestReading   data.GravityReading    `json:"latestReading"`
	Attenuation     float64                `json:"attenuation"`
	StartDate       time.Time              `json:"startDate"`
	LastUpdate      time.Time              `json:"lastUpdate"`
	Active          bool                   `json:"active"`
}

type LightweightBatch struct {
}

type BatchParam struct {
	ID         bson.ObjectId   `json:"id"`
	RecipeName string          `json:"recipe"`
	UniqueID   string          `json:"stringId"`
	Hydrometer data.Hydrometer `json:"hydrometer"`
	StartDate  time.Time       `json:"startDate"`
}

func convertDatabaseBatches(b []*data.Batch, lightweight bool) ([]*Batch, error) {
	batches := []*Batch{}

	for _, batch := range b {
		converted, err := convertDatabaseBatch(batch, lightweight)
		if err != nil {
			graviton.Logger.Warn("Batch not fully converted", zap.Error(err), zap.Any("Batch", batch))
		}
		batches = append(batches, converted)
	}

	return batches, nil
}

func convertDatabaseBatch(b *data.Batch, lightweight bool) (*Batch, error) {
	converted := &Batch{
		ID:         b.ID,
		RecipeName: b.RecipeName,
		UniqueID:   b.UniqueID,
		StartDate:  b.StartDate,
		LastUpdate: b.LastUpdate,
		Active:     b.Active,
	}

	if len(b.GravityReadings) > 0 {
		converted.LatestReading = b.GravityReadings[len(b.GravityReadings)-1]

		if !lightweight {
			converted.GravityReadings = &b.GravityReadings
		} else {
			converted.GravityReadings = &[]data.GravityReading{}
		}

		firstReading := b.GravityReadings[0]

		gravityChange := firstReading.Gravity - converted.LatestReading.Gravity
		gravityBaseline := firstReading.Gravity - 1.0

		converted.Attenuation = 1 - ((gravityBaseline - gravityChange) / gravityBaseline)
	} else {
		converted.GravityReadings = &[]data.GravityReading{}
		converted.Attenuation = 0
	}

	hydrometer, err := data.SingleHydrometer(bson.M{"_id": b.HydrometerID})

	if hydrometer != nil {
		convertedHydrometer, _ := convertDatabaseHydrometer(hydrometer)
		converted.Hydrometer = *convertedHydrometer
	}

	return converted, err
}

func convertAPIBatch(b *Batch) (*data.Batch, error) {
	return nil, nil
}

func convertBatchParam(b *BatchParam) (*data.Batch, error) {
	return nil, nil
}

type Hydrometer struct {
	ID             bson.ObjectId `json:"id,omitempty"`
	Name           string        `json:"name"`
	Description    string        `json:"description"`
	CurrentBatchID bson.ObjectId `json:"batch"`
}

type HydrometerParam struct {
	ID          bson.ObjectId `json:"id,omitempty" bson:"id,omitempty"`
	Name        string        `bson:"name" json:"name"`
	Description string        `bson:"description" json:"description"`
}

func convertDatabaseHydrometers(h []*data.Hydrometer) ([]*Hydrometer, error) {
	hydrometers := []*Hydrometer{}

	for _, hydrometer := range h {
		converted, err := convertDatabaseHydrometer(hydrometer)
		if err != nil {
			graviton.Logger.Warn("Hydrometer conversion failed", zap.Error(err), zap.Any("Hydrometer", hydrometer))
		}
		hydrometers = append(hydrometers, converted)
	}

	return hydrometers, nil
}

func convertDatabaseHydrometer(h *data.Hydrometer) (*Hydrometer, error) {
	converted := &Hydrometer{
		ID:             h.ID,
		Name:           h.Name,
		Description:    h.Description,
		CurrentBatchID: h.CurrentBatchID,
	}
	return converted, nil
}

func convertAPIHydrometer(h *Hydrometer) (*data.Hydrometer, error) {
	return nil, nil
}

func convertHydrometerParam(h *HydrometerParam) (*data.Hydrometer, error) {
	return nil, nil
}
