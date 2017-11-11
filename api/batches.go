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

	batches, err := data.QueryBatches(bson.M{})

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
	return nil
}

func EditBatch(c echo.Context) error {
	return nil
}

func AddReading(c echo.Context) error {
	return nil
}

func FinishBatch(c echo.Context) error {
	return nil
}
