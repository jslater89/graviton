package auth

import (
	"time"

	"github.com/markbates/goth"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Role struct {
	Permissions []Permission `bson:"permissions" json:"permissions"`
}

type Permission struct {
	Path     string `bson:"path" json:"path"`
	CanRead  bool   `bson:"canRead" json:"canRead"`
	CanWrite bool   `bson:"canWrite" json:"canWrite"`
}

type User struct {
	ID    bson.ObjectId   `bson:"_id"`
	Email string          `bson:"email"`
	Roles []bson.ObjectId `bson:"roles"`
}

type APIUser struct {
	ID    bson.ObjectId `json:"_id"`
	Email string        `json:"email"`
	Roles []Role        `json:"roles"`
}

type Session struct {
	ID        bson.ObjectId `bson:"_id,omitempty"`
	User      User          `bson:"user"`
	UserInfo  goth.User     `bson:"userInfo"`
	Token     string        `bson:"token"`
	CreatedAt time.Time     `bson:"created"`
	ExpiresAt time.Time     `bson:"expiry"`
}

func initLocalSessionStore(maxAge int) {
	db.sessionCollection.EnsureIndex(mgo.Index{
		Key:         []string{"created"},
		ExpireAfter: time.Second * time.Duration(maxAge),
	})
	db.sessionCollection.EnsureIndex(mgo.Index{
		Key:    []string{"token"},
		Unique: true,
	})
	db.sessionCollection.EnsureIndexKey("user")

	db.userCollection.EnsureIndex(mgo.Index{
		Key:    []string{"email"},
		Unique: true,
	})
}

func getSession(token string) (*Session, error) {
	session := &Session{}
	err := db.sessionCollection.Find(bson.M{"token": token}).One(session)

	return session, err
}

func deleteSession(token string) error {
	return db.sessionCollection.Remove(bson.M{"token": token})
}

func saveSession(session Session) error {
	if session.ID == "" {
		session.ID = bson.NewObjectId()
	}
	_, err := db.sessionCollection.UpsertId(session.ID, session)
	return err
}

func getOrCreateUser(email string) (*User, error) {
	user := &User{}
	err := db.userCollection.Find(bson.M{"email": email}).One(user)

	if err != nil && err != mgo.ErrNotFound {
		return nil, err
	} else if err == mgo.ErrNotFound {
		user.Email = email
		user.ID = bson.NewObjectId()
		_, err = db.userCollection.UpsertId(user.ID, user)

		if err != nil {
			return nil, err
		}
	}

	return user, nil
}
