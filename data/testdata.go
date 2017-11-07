package data

import (
	"fmt"
	"time"

	"github.com/jslater89/graviton"
	"gopkg.in/mgo.v2/bson"
)

const testDB = "graviton-test"

func generateTestData() {
	graviton.Init()
	InitMongo("localhost", testDB)

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
		fmt.Printf("Failed to save Flue Season: %v\n", err)
	}

	flueSeason.SetHydrometer(greenHydrometer)

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
		fmt.Printf("Failed to save Hop Forward: %v\n", err)
	}
}

func getTestObjects() (*Hydrometer, *Hydrometer, *Batch, *Batch) {
	blueHydrometer := &Hydrometer{}
	greenHydrometer := &Hydrometer{}

	flueSeason := &Batch{}
	hopForward := &Batch{}

	db.hydrometerCollection.Find(bson.M{"name": "Blue Hydrometer"}).One(blueHydrometer)
	db.hydrometerCollection.Find(bson.M{"name": "Green Hydrometer"}).One(greenHydrometer)

	db.batchCollection.Find(bson.M{"stringId": "20171101-flueseason"}).One(flueSeason)
	db.batchCollection.Find(bson.M{"stringId": "20171101-hopforward"}).One(hopForward)

	fmt.Println(blueHydrometer)
	fmt.Println(greenHydrometer)
	fmt.Println(flueSeason)
	fmt.Println(hopForward)

	return blueHydrometer, greenHydrometer, flueSeason, hopForward
}

func cleanupTestData() {
	db.dbRef.DropDatabase()
}
