package api

import (
	"bytes"
	"net/http/httptest"
	"testing"

	"gopkg.in/square/go-jose.v1/json"

	"github.com/jslater89/graviton"
	"github.com/jslater89/graviton/auth"
	"github.com/jslater89/graviton/config"
	"github.com/jslater89/graviton/data"
	"github.com/labstack/echo"
)

func TestHydrometerAPI(t *testing.T) {
	graviton.InitTest()
	auth.InitOauth(config.GetConfig().MongoAddress, config.GetConfig().GetDBName())
	data.GenerateDemoData()

	sessionID := generateTestSession()

	// --------------- 1. Test GET hydrometers by name
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/api/v1/hydrometers?name=Blue%20Hydrometer", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := QueryHydrometers(c)

	if err != nil || rec.Code != 200 {
		t.Errorf("Request failed with code %d %v\n", rec.Code, err)
	}

	hydrometers := []Hydrometer{}
	err = json.NewDecoder(rec.Body).Decode(&hydrometers)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	if len(hydrometers) == 0 || len(hydrometers) > 1 || hydrometers[0].Name != "Blue Hydrometer" {
		t.Errorf("Hydrometer response incorrect %v\n", hydrometers)
	}

	// --------------- 2. Test GET available hydrometers
	e = echo.New()
	req = httptest.NewRequest(echo.GET, "/api/v1/hydrometers/available", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = QueryAvailableHydrometers(c)

	if err != nil || rec.Code != 200 {
		t.Errorf("Request failed with code %d %v\n", rec.Code, err)
	}

	hydrometers = []Hydrometer{}
	err = json.NewDecoder(rec.Body).Decode(&hydrometers)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	if len(hydrometers) != 1 {
		t.Errorf("Hydrometer response incorrect %v\n", hydrometers)
	}

	// --------------- 3. Test POST new hydrometer
	newHydrometer := HydrometerParam{
		Name:        "Pink Hydrometer",
		Description: "A pretty pink hydrometer",
		Archived:    false,
	}
	marshaledHydrometer, err := json.Marshal(newHydrometer)

	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/api/v1/hydrometers", bytes.NewBuffer(marshaledHydrometer))
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = NewHydrometer(c)

	if err != nil || rec.Code != 200 {
		t.Errorf("Request failed with code %d %v\n", rec.Code, err)
	}

	hydrometer := Hydrometer{}
	err = json.NewDecoder(rec.Body).Decode(&hydrometer)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	// --------------- 4. Test GET single hydrometer
	newHydrometerID := hydrometer.ID

	e = echo.New()
	req = httptest.NewRequest(echo.GET, "/api/v1/hydrometers/:id", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(newHydrometerID.Hex())

	err = GetHydrometer(c)

	if err != nil || rec.Code != 200 {
		t.Errorf("Request failed with code %d %v %s\n", rec.Code, err, rec.Body.Bytes())
	}

	hydrometer = Hydrometer{}
	err = json.NewDecoder(rec.Body).Decode(&hydrometer)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	if hydrometer.Name != newHydrometer.Name || hydrometer.Description != newHydrometer.Description {
		t.Errorf("New hydrometer is wrong: %v %v\n", newHydrometer, hydrometer)
	}

	// --------------- 5. Test PUT single hydrometer
	newHydrometer = HydrometerParam{
		Name:        "Pink Hydrometer",
		Description: "A modified description",
		Archived:    false,
	}
	marshaledHydrometer, err = json.Marshal(newHydrometer)

	e = echo.New()
	req = httptest.NewRequest(echo.PUT, "/api/v1/hydrometers/:id", bytes.NewBuffer(marshaledHydrometer))
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(newHydrometerID.Hex())

	err = EditHydrometer(c)

	if err != nil || rec.Code != 200 {
		t.Errorf("Request failed with code %d %v %s\n", rec.Code, err, rec.Body.Bytes())
	}

	hydrometer = Hydrometer{}
	err = json.NewDecoder(rec.Body).Decode(&hydrometer)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	if hydrometer.Name != newHydrometer.Name || hydrometer.Description != newHydrometer.Description {
		t.Errorf("New hydrometer is wrong: %v %v\n", newHydrometer, hydrometer)
	}

	// --------------- 6. Test archive single hydrometer
	e = echo.New()
	req = httptest.NewRequest(echo.DELETE, "/api/v1/hydrometers/:id", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(newHydrometerID.Hex())

	err = ArchiveHydrometer(c)

	if err != nil || rec.Code != 200 {
		t.Errorf("Request failed with code %d %v %s\n", rec.Code, err, rec.Body.Bytes())
	}

	hydrometer = Hydrometer{}
	err = json.NewDecoder(rec.Body).Decode(&hydrometer)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	if !hydrometer.Archived {
		t.Errorf("New hydrometer is wrong: %v\n", hydrometer)
	}

	// --------------- 7. Test archive hydrometer in use
	_, greenHydrometer, _, _ := data.GetTestObjects()

	e = echo.New()
	req = httptest.NewRequest(echo.DELETE, "/api/v1/hydrometers/:id", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(greenHydrometer.ID.Hex())

	err = ArchiveHydrometer(c)

	if err != nil || rec.Code != 400 {
		t.Errorf("Request failed with code %d %v %s\n", rec.Code, err, rec.Body.Bytes())
	}

	data.CleanupTestData()
}
