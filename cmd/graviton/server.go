package main

import (
	"github.com/jslater89/graviton"
	"github.com/jslater89/graviton/api"
	"github.com/jslater89/graviton/auth"
	"github.com/jslater89/graviton/config"
	"github.com/jslater89/graviton/data"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

func main() {
	graviton.Init()
	config := config.GetConfig()

	auth.InitOauth(config.MongoAddress, config.GetDBName())
	data.InitMongo(config.MongoAddress, config.GetDBName())

	if config.DemoData {
		ensureDemoData()
	}

	e := echo.New()

	e.GET("/api/v1/auth/google/login", auth.GoogleAuthLogin)
	e.GET("/api/v1/auth/google/callback", auth.GoogleAuthCallback)
	e.GET("/api/v1/auth/logout", auth.Logout)
	e.GET("/api/v1/auth/apikey", auth.GetAPIKey)
	e.POST("/api/v1/auth/apikey/reset", auth.ResetAPIKey)
	e.GET("/api/v1/users/me", auth.GetSelf)

	e.GET("/api/v1/batches", api.QueryBatches) // returns lightweight batches: last reading and attenuation only
	e.POST("/api/v1/batches", api.NewBatch)    // takes a BatchParam

	e.GET("/api/v1/batches/:id", api.GetBatch)            // returns full batch, including all readings
	e.PUT("/api/v1/batches/:id", api.EditBatch)           // takes a BatchParam, use to start batches
	e.POST("/api/v1/batch/:id/finish", api.FinishBatch)   // sets a batch inactive, releasing its hydrometer and stopping readings
	e.POST("/api/v1/batch/:id/archive", api.ArchiveBatch) // sets a batch archived, removing it from default search results

	// Called by hydrometers; the API finds the correct
	// batch by getting the correct hydrometer.
	e.POST("/api/v1/reading", api.AddReading)

	e.GET("/api/v1/hydrometers", api.QueryHydrometers)
	e.POST("/api/v1/hydrometers", api.NewHydrometer)
	e.GET("/api/v1/hydrometers/available", api.QueryAvailableHydrometers) // gets all hydrometers not currently batch-assigned

	e.GET("/api/v1/hydrometers/:id", api.GetHydrometer)
	e.PUT("/api/v1/hydrometers/:id", api.EditHydrometer)
	e.DELETE("/api/v1/hydrometers/:id", api.ArchiveHydrometer) // sets a hydrometer archived

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		Skipper:      middleware.DefaultCORSConfig.Skipper,
		AllowOrigins: []string{"http://localhost:8080"},
		AllowMethods: middleware.DefaultCORSConfig.AllowMethods,
	}))
	e.Use(middleware.Logger())

	e.Start("localhost:10000")
}

func ensureDemoData() {
	// Generate won't insert duplicates; it'll invalid-key it up
	data.GenerateDemoData()

	b, err := data.QueryBatches(bson.M{"recipe": "Hop Forward"})
	if err != nil {
		graviton.Logger.Warn("No batches", zap.Error(err))
		return
	}

	if len(b) == 0 {
		graviton.Logger.Warn("No batches")
		return
	}

	h, err := data.QueryHydrometers(bson.M{"name": "Blue Hydrometer"})
	if err != nil {
		graviton.Logger.Warn("No hydrometers", zap.Error(err))
		return
	}

	if len(h) == 0 {
		graviton.Logger.Warn("No hydrometers")
		return
	}

	if len(b[0].GravityReadings) == 0 {
		b[0].SetHydrometer(h[0])
		data.GenerateDemoReadings(b[0])
	}
}
