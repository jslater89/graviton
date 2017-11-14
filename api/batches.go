package api

import (
	"time"

	"github.com/jslater89/graviton"
	"github.com/jslater89/graviton/auth"
	"github.com/jslater89/graviton/data"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

func QueryBatches(c echo.Context) error {
	if !auth.IsAuthorized(c, "/batches") {
		return nil
	}

	query := bson.M{
		"active":   true,
		"archived": false,
	}

	batches, err := data.QueryBatches(query)

	if err != nil {
		return c.JSON(502, bson.M{"error": "database query failed"})
	}
	responseBatches, err := convertDatabaseBatches(batches, true)

	if err != nil {
		return c.JSON(502, bson.M{"error": "database conversion failed"})
	}

	return c.JSON(200, responseBatches)
}

func GetBatch(c echo.Context) error {
	if !auth.IsAuthorized(c, "/batches") {
		return nil
	}

	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {
		return c.JSON(400, bson.M{"error": "bad object id"})
	}

	batch, err := data.SingleBatch(bson.M{"_id": bson.ObjectIdHex(id)})

	if err != nil {
		graviton.Logger.Error("Failed to query single batch", zap.String("id", id), zap.Error(err))
		return c.JSON(502, bson.M{"error": "database query failed"})
	}
	responseBatch, err := convertDatabaseBatch(batch, false)

	if err != nil {
		graviton.Logger.Error("Failed to convert single batch", zap.String("id", id), zap.Error(err))
		return c.JSON(502, bson.M{"error": "database conversion failed"})
	}

	return c.JSON(200, responseBatch)
}

func NewBatch(c echo.Context) error {
	if !auth.IsAuthorized(c, "/batches") {
		return nil
	}

	batchParam := &BatchParam{}
	err := c.Bind(batchParam)
	if err != nil {
		graviton.Logger.Warn("Invalid input", zap.Error(err))
		return c.JSON(400, bson.M{"error": "invalid input"})
	}

	batch, err := convertBatchParam(batchParam)

	if err != nil {
		graviton.Logger.Error("Invalid batch object", zap.Any("Batch", batch))
		return c.JSON(400, bson.M{"error": "invalid batch object"})
	}

	batch.ID = bson.NewObjectId()

	savedBatch, err := data.AddBatch(batch)

	if err != nil {
		graviton.Logger.Error("Failed to add batch", zap.Any("Batch", batch))
		return c.JSON(502, bson.M{"error": err.Error()})
	}

	return c.JSON(200, savedBatch)
}

func EditBatch(c echo.Context) error {
	if !auth.IsAuthorized(c, "/batches") {
		return nil
	}

	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {
		return c.JSON(400, bson.M{"error": "bad object id"})
	}

	batchParam := &BatchParam{}
	err := c.Bind(batchParam)
	if err != nil {
		graviton.Logger.Warn("Invalid input", zap.Error(err))
		return c.JSON(400, bson.M{"error": "invalid input"})
	}

	batch, err := convertBatchParam(batchParam)

	if err != nil {
		return c.JSON(400, bson.M{"error": "invalid batch object"})
	}

	batch, err = data.SingleBatch(bson.M{"_id": id})

	if err != nil {
		return c.JSON(400, bson.M{"error": err.Error()})
	}

	// mergeBatchParam saves the batch and hydrometer
	err = mergeBatchParam(batchParam, batch)

	if err != nil {
		return c.JSON(400, bson.M{"error": err.Error()})
	}

	apiBatch, err := convertDatabaseBatch(batch, false)

	if err != nil {
		return c.JSON(502, bson.M{"error": err.Error()})
	}

	return c.JSON(200, apiBatch)
}

func AddReading(c echo.Context) error {
	if !auth.IsAuthorized(c, "/reading") {
		return nil
	}

	readingParam := &HydrometerReading{}
	err := c.Bind(readingParam)

	if err != nil {
		graviton.Logger.Warn("Invalid input", zap.Error(err))
		return c.JSON(400, bson.M{"error": "invalid input"})
	}

	hydrometer, err := data.SingleHydrometer(bson.M{"name": readingParam.HydrometerName})

	if err != nil {
		graviton.Logger.Error("Error getting hydrometer for reading",
			zap.String("ID", readingParam.HydrometerName),
			zap.Error(err))
		return c.JSON(400, bson.M{"error": err.Error()})
	}

	batch, err := data.SingleBatch(bson.M{
		"_id":      hydrometer.CurrentBatchID,
		"active":   true,
		"archived": false,
	})

	if err != nil {
		graviton.Logger.Warn("Hydrometer reading received for finished batch",
			zap.String("BatchID", hydrometer.CurrentBatchID.Hex()),
			zap.Error(err))
		return c.JSON(400, bson.M{"error": "reading received for missing or completed batch"})
	}

	reading := data.GravityReading{
		Gravity:        readingParam.Gravity,
		Temperature:    readingParam.Temperature,
		BatteryVoltage: readingParam.Battery,
		Date:           time.Now(),
		Hidden:         false,
		ID:             bson.NewObjectId(),
	}
	err = batch.AddReading(reading)

	if err != nil {
		graviton.Logger.Warn("Error adding reading to batch", zap.Error(err))
		return c.JSON(502, bson.M{"error": err.Error()})
	}
	return c.JSON(200, bson.M{"status": "ok"})
}

func FinishBatch(c echo.Context) error {
	if !auth.IsAuthorized(c, "/batches") {
		return nil
	}

	batchParam := &BatchParam{}
	err := c.Bind(batchParam)

	if err != nil {
		graviton.Logger.Warn("Invalid input", zap.Error(err))
		return c.JSON(400, bson.M{"error": "invalid input"})
	}

	batch, err := data.SingleBatch(bson.M{
		"_id":      batchParam.ID,
		"active":   true,
		"archived": false,
	})

	if err != nil {
		graviton.Logger.Warn("Error getting batch",
			zap.String("BatchID", batchParam.ID.Hex()),
			zap.Error(err))
		return c.JSON(400, bson.M{"error": "reading received for missing or completed batch"})
	}

	err = batch.FinishBatch()

	if err != nil {
		graviton.Logger.Warn("Error finishing batch",
			zap.Error(err))
		return c.JSON(502, bson.M{"error": "unable to finish batch"})
	}

	return c.JSON(200, batch)
}
