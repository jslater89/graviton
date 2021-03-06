package data

import (
	"testing"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/jslater89/graviton"
)

// Batch tests happen to cover all the current hydrometer features, too.

func TestSetHydrometer(t *testing.T) {
	graviton.InitTest()
	generateTestData()

	blueHydrometer, _, flueSeason, hopForward := GetTestObjects()

	// set Flue Season to Green Hydrometer
	err := flueSeason.SetHydrometer(blueHydrometer)

	if err != nil {
		t.Errorf("Failed to set hydrometer: %s", err.Error())
	}

	if flueSeason.HydrometerID != blueHydrometer.ID {
		t.Errorf("Batch hydrometer ID wrong")
	}

	if blueHydrometer.CurrentBatchID != flueSeason.ID {
		t.Errorf("Hydrometer batch ID wrong")
	}

	blueHydrometer, greenHydrometer, flueSeason, hopForward := GetTestObjects()

	if greenHydrometer.CurrentBatchID != graviton.EmptyID() {
		t.Errorf("Replaced hydrometer batch ID not reset")
	}

	// try to set Hop Forward to Green Hydrometer
	err = hopForward.SetHydrometer(blueHydrometer)

	if err == nil {
		t.Errorf("Stole another batch's hydrometer")
	}

	err = hopForward.SetHydrometerID(greenHydrometer.ID)

	if err != nil {
		t.Errorf("Unable to set hydrometer by ID")
	}

	CleanupTestData()
}

func TestAddBatch(t *testing.T) {
	graviton.InitTest()
	generateTestData()

	batch := &Batch{
		RecipeName: "Smoke on the Lauter",
		StartDate:  time.Now(),
		UniqueID:   "20171101-flueseason",
		Active:     true,
	}
	_, err := AddBatch(batch)

	if err == nil {
		t.Errorf("Added a batch with a duplicate string ID")
	}

	batch.UniqueID = "20171101-rauchbier"
	addedBatch, err := AddBatch(batch)

	if err != nil {
		t.Errorf("Unable to add new batch: %v\n", err)
	}

	blueHydrometer, greenHydrometer, _, _ := GetTestObjects()

	err = addedBatch.SetHydrometer(greenHydrometer)

	if err == nil {
		t.Errorf("Stole another batch's hydrometer")
	}

	err = addedBatch.SetHydrometer(blueHydrometer)

	if err != nil {
		t.Errorf("Unable to assign unused hydrometer: %v\n", err)
	}

	CleanupTestData()
}

func TestAddGravityReading(t *testing.T) {
	graviton.InitTest()
	generateTestData()

	_, _, flueSeason, _ := GetTestObjects()

	newReadingID := bson.NewObjectId()
	flueSeason.AddReading(GravityReading{
		ID:             newReadingID,
		BatteryVoltage: 3.7,
		Date:           time.Now().Add(time.Minute * 15),
		Gravity:        1.076,
		Temperature:    68.9,
	})

	for i, reading := range flueSeason.GravityReadings {
		if i == 1 {
			if reading.ID != newReadingID {
				t.Errorf("Reading not sorted by time: id %v in %v\n", newReadingID, flueSeason.GravityReadings)
			}
			break
		}
	}

	newReadingID = bson.NewObjectId()
	flueSeason.AddReading(GravityReading{
		ID:             newReadingID,
		BatteryVoltage: 3.7,
		Date:           time.Now().Add(time.Minute * 120),
		Gravity:        1.064,
		Temperature:    71.2,
	})

	reading := flueSeason.GravityReadings[len(flueSeason.GravityReadings)-1]

	if reading.ID != newReadingID {
		t.Errorf("Reading not sorted by time: id %v in %v\n", newReadingID, flueSeason.GravityReadings)
	}

	CleanupTestData()
}

func TestHideGravityReading(t *testing.T) {
	graviton.InitTest()
	generateTestData()

	_, _, flueSeason, _ := GetTestObjects()

	for i, reading := range flueSeason.GravityReadings {
		if i == 1 {
			flueSeason.HideReadingID(reading.ID)
			break
		}
	}

	for i, reading := range flueSeason.GravityReadings {
		if i == 1 {
			if !reading.Hidden {
				t.Errorf("Not correctly hidden")
				break
			}
		}
	}

	CleanupTestData()
}

func TestFinishBatch(t *testing.T) {
	graviton.InitTest()
	generateTestData()

	_, greenHydrometer, flueSeason, hopForward := GetTestObjects()

	err := flueSeason.FinishBatch()
	if err != nil {
		t.Errorf("Error finishing batch")
	}

	_, greenHydrometer, flueSeason, hopForward = GetTestObjects()

	if flueSeason.HydrometerID != greenHydrometer.ID {
		t.Errorf("Didn't keep a record of hydrometer used")
	}

	if greenHydrometer.CurrentBatchID != graviton.EmptyID() {
		t.Errorf("Didn't clear hydrometer in use flag")
	}

	err = hopForward.SetHydrometer(greenHydrometer)

	if err != nil {
		t.Errorf("Unable to set hydrometer from finished batch to new batch")
	}

	CleanupTestData()
}
