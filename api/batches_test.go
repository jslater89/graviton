package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jslater89/graviton"
	"github.com/jslater89/graviton/auth"
	"github.com/jslater89/graviton/config"
	"github.com/jslater89/graviton/data"
	"github.com/labstack/echo"
	"github.com/markbates/goth"
	"gopkg.in/mgo.v2/bson"
)

func TestBatchAPI(t *testing.T) {
	graviton.InitTest()
	auth.InitOauth(config.GetConfig().MongoAddress, config.GetConfig().GetDBName())
	data.GenerateDemoData()

	sessionID := generateTestSession()

	// ---------------- 1. Test GET batches by name
	e := echo.New()
	req := httptest.NewRequest(echo.GET, "/api/v1/batches?recipe=Flue%20Season", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := QueryBatches(c)

	if err != nil || rec.Code != 200 {
		t.Errorf("Request failed with code %d %v\n", rec.Code, err)
	}

	batches := []Batch{}
	err = json.NewDecoder(rec.Body).Decode(&batches)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	if len(batches) == 0 || len(batches) > 1 || batches[0].UniqueID != "20171101-flueseason" {
		t.Errorf("Batch response incorrect %v\n", batches)
	}

	flueSeason := batches[0]

	// ---------------- 2. Test GET batch by ID
	e = echo.New()
	req = httptest.NewRequest(echo.GET, "/api/v1/batches/:id", nil)
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(flueSeason.ID.Hex())

	err = GetBatch(c)

	if err != nil || rec.Code != 200 {
		t.Errorf("Request failed with code %d %v\n", rec.Code, err)
	}

	receivedBatch := &Batch{}
	err = json.NewDecoder(rec.Body).Decode(&receivedBatch)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	if receivedBatch.UniqueID != flueSeason.UniqueID {
		t.Errorf("Mismatch result from individual get")
	}

	// ---------------- 3. Test PUT existing batch, including setting hydrometer to nil
	greenHydrometerID := bson.ObjectIdHex(flueSeason.Hydrometer.ID.Hex())
	flueSeason.Hydrometer.ID = ""

	body, _ := json.Marshal(flueSeason)

	e = echo.New()
	req = httptest.NewRequest(echo.PUT, "/api/v1/batches/:id", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(flueSeason.ID.Hex())

	err = EditBatch(c)

	if err != nil || rec.Code != 200 {
		bodyBytes := rec.Body.Bytes()
		t.Errorf("Request failed with code %d %v %s\n", rec.Code, err, string(bodyBytes))
	}

	receivedBatch = &Batch{}
	err = json.NewDecoder(rec.Body).Decode(&receivedBatch)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	if receivedBatch.Hydrometer.ID != "" {
		t.Errorf("Unable to unset hydrometer")
	}

	// ---------------- 4. Test POST new batch, setting hydrometer to old hydrometer
	batchParam := BatchParam{
		Active:   true,
		Archived: false,
		Hydrometer: Hydrometer{
			ID: greenHydrometerID,
		},
		RecipeName: "'Tis the Saison",
		UniqueID:   "20171101-saison",
	}

	body, _ = json.Marshal(batchParam)

	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/api/v1/batches", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = NewBatch(c)

	if err != nil || rec.Code != 200 {
		bodyBytes := rec.Body.Bytes()
		t.Errorf("Request failed with code %d %v %s\n", rec.Code, err, string(bodyBytes))
	}

	receivedBatch = &Batch{}
	err = json.NewDecoder(rec.Body).Decode(&receivedBatch)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	if receivedBatch.Hydrometer.ID == "" {
		fmt.Println(receivedBatch)
		t.Errorf("New batch hydrometer setting didn't work")
	}

	tisTheSaison := *receivedBatch

	// ---------------- 5. Test POST new batch with no hydrometer set
	batchParam = BatchParam{
		Active:     false,
		Archived:   true,
		RecipeName: "Pound Foolish",
		UniqueID:   "20171101-poundfoolish",
	}

	body, _ = json.Marshal(batchParam)

	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/api/v1/batches", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = NewBatch(c)

	if err != nil || rec.Code != 200 {
		bodyBytes := rec.Body.Bytes()
		t.Errorf("Request failed with code %d %v %s\n", rec.Code, err, string(bodyBytes))
	}

	receivedBatch = &Batch{}
	err = json.NewDecoder(rec.Body).Decode(&receivedBatch)

	if err != nil {
		t.Errorf("Unable to decode response: %v\n", err)
	}

	if receivedBatch.Active == true || receivedBatch.Archived == true {
		t.Errorf("Boolean parameters set incorrectly")
	}

	// ---------------- 6. Test add reading
	readingParam := HydrometerReading{
		Battery:        3.6,
		Gravity:        1.074,
		Temperature:    68.8,
		HydrometerName: "Green Hydrometer",
	}

	body, _ = json.Marshal(readingParam)

	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/api/v1/reading", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	err = AddReading(c)

	if err != nil || rec.Code != 200 {
		bodyBytes := rec.Body.Bytes()
		t.Errorf("Request failed with code %d %v %s\n", rec.Code, err, string(bodyBytes))
	}

	dataBatch, _ := data.SingleBatch(bson.M{"_id": tisTheSaison.ID})

	if len(dataBatch.GravityReadings) != 1 {
		t.Errorf("Incorrect number of readings")
	}

	// ---------------- 7. Test finish batch
	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/api/v1/batches/:id/finish", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tisTheSaison.ID.Hex())

	err = FinishBatch(c)

	if err != nil || rec.Code != 200 {
		bodyBytes := rec.Body.Bytes()
		t.Errorf("Request failed with code %d %v %s\n", rec.Code, err, string(bodyBytes))
	}

	dataBatch, _ = data.SingleBatch(bson.M{"_id": tisTheSaison.ID})

	if dataBatch.Active {
		t.Errorf("Batch still active")
	}

	// ---------------- 8. Test finish batch
	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/api/v1/batches/:id/archive", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+sessionID.Hex())
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)
	c.SetParamNames("id")
	c.SetParamValues(tisTheSaison.ID.Hex())

	err = ArchiveBatch(c)

	if err != nil || rec.Code != 200 {
		bodyBytes := rec.Body.Bytes()
		t.Errorf("Request failed with code %d %v %s\n", rec.Code, err, string(bodyBytes))
	}

	dataBatch, _ = data.SingleBatch(bson.M{"_id": tisTheSaison.ID})

	if !dataBatch.Archived {
		t.Errorf("Batch not archived")
	}

	data.CleanupTestData()
}

func generateTestSession() bson.ObjectId {
	// G
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", strings.NewReader("{}"))
	//req.Header.Set("Authorization", "Bearer abcdefabcdefabcdefabcdef")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	user := goth.User{
		Email:     "testuser@example.com",
		ExpiresAt: time.Now().Add(60 * time.Second),
	}

	sessionID := auth.HandleUser(c, user)
	return sessionID
}
