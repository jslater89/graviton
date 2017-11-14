package data

import (
	"github.com/jslater89/graviton/config"
	"gopkg.in/mgo.v2"
)

type database struct {
	session              *mgo.Session
	dbRef                *mgo.Database
	batchCollection      *mgo.Collection
	hydrometerCollection *mgo.Collection
}

var db database

func InitMongo(dbAddr string, database string) error {
	var err error
	db.session, err = mgo.Dial("localhost")

	if err != nil {
		return err
	}

	db.dbRef = db.session.DB(config.GetConfig().GetDBName())
	db.batchCollection = db.dbRef.C("batches")
	db.hydrometerCollection = db.dbRef.C("hydrometers")

	ensureIndices()

	return nil
}

func ensureIndices() {
	db.batchCollection.EnsureIndexKey("recipe")
	db.batchCollection.EnsureIndexKey("-startDate")
	db.batchCollection.EnsureIndexKey("-lastUpdate")
	db.batchCollection.EnsureIndexKey("hydrometer")
	db.batchCollection.EnsureIndexKey("active")
	db.batchCollection.EnsureIndex(mgo.Index{
		Key:    []string{"stringId"},
		Unique: true,
	})

	db.hydrometerCollection.EnsureIndex(mgo.Index{
		Key:    []string{"name"},
		Unique: true,
	})
}
