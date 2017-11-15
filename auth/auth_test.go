package auth

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/jslater89/graviton"
	"github.com/jslater89/graviton/config"
	"github.com/jslater89/graviton/data"
	"github.com/labstack/echo"
	"github.com/markbates/goth"
	"gopkg.in/mgo.v2/bson"
)

func TestExtractBearer(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer abcdef")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	bearer := extractBearer(c)

	if bearer != "abcdef" {
		t.Errorf("Failed to get bearer from header")
	}

	e = echo.New()
	req = httptest.NewRequest(echo.POST, "/", strings.NewReader("{}"))
	req.AddCookie(getCookie("abcdef"))
	rec = httptest.NewRecorder()
	c = e.NewContext(req, rec)

	bearer = extractBearer(c)

	if bearer != "abcdef" {
		t.Errorf("Failed to get bearer from cookie")
	}
}

func TestHandleUser(t *testing.T) {
	generateTestData()

	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", strings.NewReader("{}"))
	req.AddCookie(getCookie("abcdefabcdefabcdefabcdef"))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	user := goth.User{
		Email:     "testuser@mail.com",
		ExpiresAt: time.Now().Add(30 * time.Second),
	}

	sessionID := HandleUser(c, user)

	if sessionID == graviton.EmptyID() {
		t.Errorf("Failed to make session for new user")
	}

	if n, _ := db.userCollection.Find(bson.M{}).Count(); n == 0 {
		t.Errorf("Failed to create new user")
	}

	sess, err := getSession("abcdefabcdefabcdefabcdef")

	if err != nil {
		t.Errorf("Couldn't get session: %v", err)
	}

	t.Logf("Session: %v", sess)

	if !IsAuthorized(c, "/") {
		t.Errorf("User can't write")
	}

	req = httptest.NewRequest(echo.GET, "/", nil)
	req.AddCookie(getCookie("abcdefabcdefabcdefabcdef"))
	c = e.NewContext(req, rec)

	if !IsAuthorized(c, "/") {
		t.Errorf("User can't read")
	}

	cleanupTestData()
}

func TestAuthorizeAPIKey(t *testing.T) {
	generateTestData()

	e := echo.New()
	req := httptest.NewRequest(echo.POST, "/", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer "+getOrCreateAPIKey(false))
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if !IsAuthorized(c, "/arbitrary/path") {
		t.Errorf("API key auth not successful")
	}

	db.mongoDB.DropDatabase()
}

func generateTestData() {
	graviton.InitTest()
	data.GenerateDemoData()
	InitOauth(config.GetConfig().MongoAddress, config.GetConfig().GetDBName())
}

func cleanupTestData() {
	db.mongoDB.DropDatabase()
}
