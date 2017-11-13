package api

import (
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
	batchParam := &BatchParam{}
	c.Bind(batchParam)

	batch, err := convertBatchParam(batchParam)

	if err != nil {
		return c.JSON(400, bson.M{"error": "invalid batch object"})
	}

	if batch.ID == "" {
		batch.ID = bson.NewObjectId()
	}

	savedBatch, err := data.AddBatch(batch)

	if err != nil {
		return c.JSON(400, bson.M{"error": err})
	}

	return c.JSON(200, savedBatch)
}

func EditBatch(c echo.Context) error {
	batchParam := &BatchParam{}
	c.Bind(batchParam)

	batch, err := convertBatchParam(batchParam)

	if err != nil {
		return c.JSON(400, bson.M{"error": "invalid batch object"})
	}

	batch, err = data.SingleBatch(bson.M{"_id": batch.ID})

	if err != nil {
		return c.JSON(400, bson.M{"error": err})
	}

	err = mergeBatchParam(batchParam, batch)

	if err != nil {
		return c.JSON(400, bson.M{"error": err})
	}

	apiBatch, err := convertDatabaseBatch(batch, false)

	if err != nil {
		return c.JSON(502, bson.M{"error": err})
	}

	return c.JSON(200, apiBatch)
}

func AddReading(c echo.Context) error {
	readingParam := &HydrometerReading{}
	c.Bind(readingParam)

	return c.JSON(200, bson.M{"status": "ok"})
}

func FinishBatch(c echo.Context) error {
	return nil
}
