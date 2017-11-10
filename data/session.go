package data

import (
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"gopkg.in/mgo.v2/bson"
)

type Permission struct {
	Path     string `bson:"path" json:"path"`
	CanRead  bool   `bson:"canRead" json:"canRead"`
	CanWrite bool   `bson:"canWrite" json:"canWrite"`
}

type User struct {
	Sessions    []Session    `bson:"sessions"`
	Permissions []Permission `bson:"permissions"`
}

type Session struct {
	ID        bson.ObjectId `bson:"_id"`
	User      User          `bson:"user"`
	JWT       jwt.Token     `bson:"token"`
	ExpiresAt time.Time     `bson:"expiry"`
}
