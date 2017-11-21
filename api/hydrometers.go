package api

import (
	"github.com/jslater89/graviton"
	"github.com/jslater89/graviton/data"
	"github.com/labstack/echo"
	"gopkg.in/mgo.v2/bson"
)

func QueryHydrometers(c echo.Context) error {
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
	return nil
}

func GetHydrometer(c echo.Context) error {
	return nil
}

func EditHydrometer(c echo.Context) error {
	return nil
}

func ArchiveHydrometer(c echo.Context) error {
	return nil
}

func parseHydrometerQuery(c echo.Context, query bson.M) {
	return
}
