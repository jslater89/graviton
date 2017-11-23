package api

import (
	"github.com/jslater89/graviton"
	"github.com/jslater89/graviton/auth"
	"github.com/jslater89/graviton/data"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

func QueryHydrometers(c echo.Context) error {
	if !auth.IsAuthorized(c, "/hydrometers") {
		return nil
	}

	query := bson.M{
		"archived": false,
	}

	parseHydrometerQuery(c, query)

	hydrometers, err := data.QueryHydrometers(bson.M{})

	if err != nil {
		c.String(502, "database query failed")
		return err
	}

	responseHydrometers, err := convertDatabaseHydrometers(hydrometers)
	if err != nil {
		return c.JSON(502, bson.M{"error": "database conversion failed"})
	}

	return c.JSON(200, responseHydrometers)
}

func QueryAvailableHydrometers(c echo.Context) error {
	if !auth.IsAuthorized(c, "/hydrometers") {
		return nil
	}

	hydrometers, err := data.QueryHydrometers(bson.M{"batch": graviton.EmptyID(), "archived": false})

	if err != nil {
		c.String(502, "database query failed")
		return err
	}

	responseHydrometers, err := convertDatabaseHydrometers(hydrometers)
	if err != nil {
		return c.JSON(502, bson.M{"error": "database conversion failed"})
	}

	return c.JSON(200, responseHydrometers)
}

func NewHydrometer(c echo.Context) error {
	if !auth.IsAuthorized(c, "/hydrometers") {
		return nil
	}

	hydrometerParam := &HydrometerParam{}
	err := c.Bind(hydrometerParam)

	if err != nil {
		return defaultErrorResponse(c, 400, err)
	}

	databaseHydrometer, err := convertHydrometerParam(hydrometerParam)

	if err != nil {
		return defaultErrorResponse(c, 502, err)
	}

	err = databaseHydrometer.Save()

	if err != nil {
		return defaultErrorResponse(c, 502, err)
	}

	responseHydrometer, err := convertDatabaseHydrometer(databaseHydrometer)

	if err != nil {
		return defaultErrorResponse(c, 502, err)
	}
	return c.JSON(200, responseHydrometer)
}

func GetHydrometer(c echo.Context) error {
	if !auth.IsAuthorized(c, "/hydrometers") {
		return nil
	}

	id := c.Param("id")

	if !bson.IsObjectIdHex(id) {
		return c.JSON(400, bson.M{"error": "bad object id"})
	}

	hydrometer, err := data.SingleHydrometer(bson.M{"_id": bson.ObjectIdHex(id)})

	if err != nil {
		graviton.Logger.Error("Failed to query single hydrometer", zap.String("id", id), zap.Error(err))
		return c.JSON(502, bson.M{"error": "database query failed"})
	}

	responseHydrometer, err := convertDatabaseHydrometer(hydrometer)

	if err != nil {
		graviton.Logger.Error("Failed to convert single hydrometer", zap.String("id", id), zap.Error(err))
		return c.JSON(502, bson.M{"error": "database conversion failed"})
	}

	return c.JSON(200, responseHydrometer)
}

func EditHydrometer(c echo.Context) error {
	if !auth.IsAuthorized(c, "/hydrometers") {
		return nil
	}

	return nil
}

func ArchiveHydrometer(c echo.Context) error {
	if !auth.IsAuthorized(c, "/hydrometers") {
		return nil
	}

	return nil
}

func parseHydrometerQuery(c echo.Context, query bson.M) {
	nameParam := c.QueryParam("name")

	if nameParam != "" {
		query["name"] = nameParam
	}

	return
}
