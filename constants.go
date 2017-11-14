package graviton

import (
	"github.com/jslater89/graviton/config"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

var Logger *zap.Logger

var emptyID = bson.ObjectIdHex("2a2a2a2a2a2a2a2a2a2a2a2a")

func Init() error {
	config.Load(nil)

	var err error
	Logger, err = zap.NewProduction()

	if err != nil {
		return err
	}

	return nil
}

func InitTest() error {
	config.Load(nil)
	config.OverrideTest(true)

	var err error
	Logger, err = zap.NewDevelopment()

	if err != nil {
		return err
	}

	return nil
}

func EmptyID() bson.ObjectId {
	return emptyID
}
