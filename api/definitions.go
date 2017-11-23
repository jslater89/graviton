package api

import (
	"time"

	"gopkg.in/mgo.v2"

	"github.com/jslater89/graviton"
	"github.com/jslater89/graviton/data"
	"github.com/labstack/echo"
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
	ABV             float64                `json:"abv"`
	StartDate       time.Time              `json:"startDate"`
	LastUpdate      time.Time              `json:"lastUpdate"`
	Active          bool                   `json:"active"`
	Archived        bool                   `json:"archived"`
}

type LightweightBatch struct {
}

type BatchParam struct {
	ID         bson.ObjectId `json:"id"`
	RecipeName string        `json:"recipe"`
	UniqueID   string        `json:"stringId"`
	Hydrometer Hydrometer    `json:"hydrometer"`
	StartDate  time.Time     `json:"startDate"`
	Active     bool          `json:"active"`
	Archived   bool          `json:"archived"`
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
		Archived:   b.Archived,
	}

	if len(b.GravityReadings) > 0 {
		converted.LatestReading = b.GravityReadings[len(b.GravityReadings)-1]

		if !lightweight {
			converted.GravityReadings = &b.GravityReadings
		} else {
			converted.GravityReadings = &[]data.GravityReading{}
		}

		firstReading := b.GravityReadings[0]
		currentReading := converted.LatestReading

		og := firstReading.Gravity
		cg := currentReading.Gravity

		gravityChange := og - cg
		gravityBaseline := og - 1.0

		converted.Attenuation = 1 - ((gravityBaseline - gravityChange) / gravityBaseline)
		converted.ABV = ((76.08 * (og - cg) / (1.775 - og)) * (cg / 0.794))
	} else {
		converted.GravityReadings = &[]data.GravityReading{}
		converted.Attenuation = 0
		converted.ABV = 0
	}

	hydrometer, err := data.SingleHydrometer(bson.M{"_id": b.HydrometerID})

	if err != nil && err == mgo.ErrNotFound {
		converted.Hydrometer = Hydrometer{}
		err = nil
	} else if hydrometer != nil {
		convertedHydrometer, _ := convertDatabaseHydrometer(hydrometer)
		converted.Hydrometer = *convertedHydrometer
	}

	return converted, err
}

func convertBatchParam(b *BatchParam) (*data.Batch, error) {
	batch := &data.Batch{
		ID:           b.ID,
		HydrometerID: b.Hydrometer.ID,
		RecipeName:   b.RecipeName,
		StartDate:    b.StartDate,
		UniqueID:     b.UniqueID,
		Active:       b.Active,
		Archived:     false,
	}
	return batch, nil
}

func mergeBatchParam(param *BatchParam, batch *data.Batch) error {
	batch.RecipeName = param.RecipeName
	batch.StartDate = param.StartDate
	batch.UniqueID = param.UniqueID
	batch.Active = param.Active
	batch.Archived = param.Archived

	if param.Hydrometer.ID != "" && param.Hydrometer.ID != graviton.EmptyID() {
		hydrometer, err := data.SingleHydrometer(bson.M{"_id": param.Hydrometer.ID})

		if err != nil {
			return err
		}

		return batch.SetHydrometer(hydrometer)
	}

	return batch.SetHydrometer(&data.Hydrometer{
		ID: graviton.EmptyID(),
	})
}

type HydrometerReading struct {
	HydrometerName string  `json:"name"`
	Gravity        float64 `json:"gravity"`
	Temperature    float64 `json:"temperature"`
	Battery        float64 `json:"battery"`
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

func convertHydrometerParam(h *HydrometerParam) (*data.Hydrometer, error) {
	hydrometer := &data.Hydrometer{
		ID:          h.ID,
		Name:        h.Name,
		Description: h.Description,
	}
	return hydrometer, nil
}

func defaultErrorResponse(c echo.Context, code int, err error) error {
	return c.JSON(code, bson.M{"error": err.Error()})
}
