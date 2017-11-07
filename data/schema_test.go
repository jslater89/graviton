package data

import (
	"testing"
)

func TestSetHydrometer(t *testing.T) {
	generateTestData()

	_, greenHydrometer, flueSeason, hopForward := getTestObjects()

	// set Flue Season to Green Hydrometer
	err := flueSeason.SetHydrometer(greenHydrometer)

	if err != nil {
		t.Errorf("Failed to set hydrometer: %s", err.Error())
	}

	if flueSeason.HydrometerID != greenHydrometer.ID {
		t.Errorf("Batch hydrometer ID wrong")
	}

	if greenHydrometer.CurrentBatchID != flueSeason.ID {
		t.Errorf("Hydrometer batch ID wrong")
	}

	// try to set Hop Forward to Green Hydrometer
	err = hopForward.SetHydrometer(greenHydrometer)

	if err == nil {
		t.Errorf("Assigned two hydrometers to one batch")
	}

	cleanupTestData()
}

func TestAddBatch(t *testing.T) {

}

func TestAddGravityReading(t *testing.T) {

}

func TestHideGravityReading(t *testing.T) {

}

func TestFinishBatch(t *testing.T) {

}
