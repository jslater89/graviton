package data

import (
	"fmt"
	"time"

	"math/rand"

	"github.com/jslater89/graviton"
	"github.com/jslater89/graviton/config"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

// GenerateDemoData populates the database with some
// example information.
func GenerateDemoData() {
	generateTestData()
}

// GenerateDemoReadings generates a full time series of
// readings for the given batch, spaced a half hour apart.
func GenerateDemoReadings(b *Batch) {
	startingGravity := rand.Float64()*0.07 + 1.04
	finalGravity := 1.0 + 0.25*(startingGravity-1.0)

	gravityReadingCount := 2*72 + rand.Intn(72)
	averageGravityDelta := (startingGravity - finalGravity) / float64(gravityReadingCount)

	currentGravity := startingGravity
	for i := 0; i < gravityReadingCount; i++ {
		reading := GravityReading{
			BatteryVoltage: 3.4,
			Gravity:        currentGravity,
			Date:           time.Now().Add(time.Duration(30*i) * time.Minute),
			Temperature:    68.5 + rand.Float64()*3 - 1,
		}

		b.AddReading(reading)
		currentGravity -= averageGravityDelta + rand.Float64()*0.001 - 0.0005
	}
}

func generateTestData() {
	InitMongo("localhost", config.GetConfig().GetDBName())

	blueHydrometer := &Hydrometer{
		ID:          bson.NewObjectId(),
		Name:        "Blue Hydrometer",
		Description: "A fake hydrometer with blue plastic.",
		// test auto-fill current batch ID
	}
	greenHydrometer := &Hydrometer{
		ID:             bson.NewObjectId(),
		Name:           "Green Hydrometer",
		Description:    "A fake hydrometer with green plastic.",
		CurrentBatchID: graviton.EmptyID(),
	}

	blueHydrometer.Save()
	greenHydrometer.Save()

	flueSeason := &Batch{
		ID:           bson.NewObjectId(),
		Active:       true,
		RecipeName:   "Flue Season",
		UniqueID:     "20171101-flueseason",
		StartDate:    time.Now(),
		HydrometerID: graviton.EmptyID(),
	}
	err := flueSeason.Save()

	if err != nil {
		graviton.Logger.Error("Error saving demo data", zap.Error(err))
	}

	err = flueSeason.SetHydrometer(greenHydrometer)

	if err != nil {
		graviton.Logger.Error("Error saving demo data", zap.Error(err))
	}

	flueSeason.AddReading(GravityReading{
		Date:           time.Now(),
		Gravity:        1.077,
		Temperature:    68.8,
		BatteryVoltage: 3.7,
		Hidden:         false,
	})
	flueSeason.AddReading(GravityReading{
		Date:           time.Now().Add(time.Minute * 30),
		Gravity:        1.075,
		Temperature:    68.1,
		BatteryVoltage: 3.69,
		Hidden:         false,
	})
	flueSeason.AddReading(GravityReading{
		Date:           time.Now().Add(time.Minute * 60),
		Gravity:        1.074,
		Temperature:    69.5,
		BatteryVoltage: 3.68,
		Hidden:         false,
	})
	flueSeason.AddReading(GravityReading{
		Date:           time.Now().Add(time.Minute * 90),
		Gravity:        1.071,
		Temperature:    71.1,
		BatteryVoltage: 3.67,
		Hidden:         false,
	})

	hopForward := &Batch{
		ID:           bson.NewObjectId(),
		Active:       true,
		RecipeName:   "Hop Forward",
		UniqueID:     "20171101-hopforward",
		StartDate:    time.Now(),
		HydrometerID: graviton.EmptyID(),
	}
	err = hopForward.Save()

	if err != nil {
		graviton.Logger.Error("Error saving demo data", zap.Error(err))
	}
}

func getTestObjects() (*Hydrometer, *Hydrometer, *Batch, *Batch) {
	blueHydrometer := &Hydrometer{}
	greenHydrometer := &Hydrometer{}

	flueSeason := &Batch{}
	hopForward := &Batch{}

	hydrometers := []*Hydrometer{}
	batches := []*Batch{}

	hydrometers, _ = QueryHydrometers(bson.M{"name": "Blue Hydrometer"})
	blueHydrometer = hydrometers[0]

	greenHydrometer, _ = SingleHydrometer(bson.M{"name": "Green Hydrometer"})

	batches, _ = QueryBatches(bson.M{"stringId": "20171101-flueseason"})
	flueSeason = batches[0]

	hopForward, _ = SingleBatch(bson.M{"stringId": "20171101-hopforward"})

	fmt.Println(blueHydrometer)
	fmt.Println(greenHydrometer)
	fmt.Println(flueSeason)
	fmt.Println(hopForward)

	return blueHydrometer, greenHydrometer, flueSeason, hopForward
}

func CleanupTestData() {
	db.dbRef.DropDatabase()
}
