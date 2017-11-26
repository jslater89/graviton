package auth

import (
	"crypto/rand"
	"net/http"
	"time"

	"github.com/jslater89/graviton"
	"go.uber.org/zap"

	"github.com/labstack/echo"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/jslater89/graviton/config"
	"github.com/kidstuff/mongostore"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"
)

type database struct {
	session           *mgo.Session
	mongoDB           *mgo.Database
	gothicCollection  *mgo.Collection
	sessionCollection *mgo.Collection
	userCollection    *mgo.Collection
	roleCollection    *mgo.Collection
	apiKeyCollection  *mgo.Collection
	mongoStore        *mongostore.MongoStore // Only for gothic
}

var db database

func GetStore() *mongostore.MongoStore {
	return db.mongoStore
}

func InitOauth(dbAddress string, dbName string) {
	config := config.GetConfig()

	url := ""
	if config.UseSSL {
		url = "https://"
	} else {
		url = "http://"
	}

	url += config.ServerAddress + "/api/v1/auth/google/callback"

	goth.UseProviders(gplus.New(config.GoogleClientID, config.GoogleSecret, url))
	gothic.GetProviderName = func(*http.Request) (string, error) {
		return "gplus", nil
	}

	var err error
	db.session, err = mgo.Dial(dbAddress)
	if err != nil {
		panic(err)
	}
	db.mongoDB = db.session.DB(dbName)

	db.gothicCollection = db.mongoDB.C("oauth_store")
	db.sessionCollection = db.mongoDB.C("sessions")
	db.userCollection = db.mongoDB.C("users")
	db.roleCollection = db.mongoDB.C("roles")
	db.apiKeyCollection = db.mongoDB.C("apikey")

	initLocalSessionStore(3600)
	verifyBaseRoles()

	store := mongostore.NewMongoStore(db.gothicCollection, 300, true, []byte("secret-key"))

	db.mongoStore = store
	gothic.Store = store
}

func HandleUser(c echo.Context, user goth.User) bson.ObjectId {
	var sess *Session
	var err error
	var sessionID bson.ObjectId
	sessionString := extractBearer(c)

	if !bson.IsObjectIdHex(sessionString) {
		graviton.Logger.Info("Bad auth header")
		sessionID = bson.ObjectId(generateSessionToken())
	} else {
		sessionID = bson.ObjectIdHex(sessionString)
		sess, err = getSession(sessionID.Hex())
	}

	if sess == nil || err != nil {
		sess = &Session{Token: sessionID.Hex()}
		graviton.Logger.Info("Created new session for user", zap.String("Email", user.Email), zap.String("Session ID", sessionID.Hex()))
	}

	localUser, err := getOrCreateUser(user.Email)

	if err != nil {
		graviton.Logger.Error("Unable to get user", zap.Error(err))
		return graviton.EmptyID()
	}

	sess.CreatedAt = time.Now()
	sess.ExpiresAt = user.ExpiresAt
	sess.UserInfo = user
	sess.User = *localUser

	if err != nil {
		graviton.Logger.Warn("Error fetching session", zap.Any("User", user), zap.Error(err))
	}

	err = saveSession(*sess)

	if err != nil {
		graviton.Logger.Warn("Error storing session", zap.Any("User", user), zap.Error(err))
	}
	return sessionID
}

func generateSessionToken() string {
	secureBytes := make([]byte, 12)
	rand.Read(secureBytes)
	return string(secureBytes)
}
