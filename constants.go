package graviton

import (
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

var Logger *zap.Logger

var emptyID = bson.ObjectIdHex("2a2a2a2a2a2a2a2a2a2a2a2a")

func Init() error {
	var err error
	Logger, err = zap.NewProduction()

	if err != nil {
		return err
	}

	return nil
}

func EmptyID() bson.ObjectId {
	return emptyID
}
