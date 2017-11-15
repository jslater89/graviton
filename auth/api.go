package auth

import (
	"net/http"
	"time"

	"github.com/jslater89/graviton"
	"go.uber.org/zap"

	"github.com/labstack/echo"
	"github.com/markbates/goth/gothic"
	"gopkg.in/mgo.v2/bson"
)

func GoogleAuthLogin(c echo.Context) error {
	// try to get the user without re-authenticating
	if token := extractBearer(c); token != "" {
		session, err := getSession(token)
		if err == nil && checkSessionExpiration(session) {
			c.SetCookie(getCookie(session.Token))
			return c.Redirect(307, "http://localhost:8080/#/authenticated?bearer="+session.Token)
		}
	}

	gothic.BeginAuthHandler(c.Response(), c.Request())

	return c.JSON(200, bson.M{"status": "ok"})
}

func GoogleAuthCallback(c echo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		graviton.Logger.Error("OAuth callback returned error", zap.Error(err))
		return c.JSON(502, bson.M{"error": err.Error()})
	}

	sessionID := HandleUser(c, user)

	if sessionID == graviton.EmptyID() {
		return c.JSON(502, bson.M{"error": "could not fetch user"})
	}
	c.SetCookie(getCookie(sessionID.Hex()))
	return c.Redirect(307, "http://localhost:8080/#/authenticated?bearer="+sessionID.Hex())
}

func GetSelf(c echo.Context) error {
	token := extractBearer(c)

	if token == "" || !bson.IsObjectIdHex("token") {
		return c.JSON(400, bson.M{"error": "invalid session"})
	}

	session, err := getSession(token)

	if err != nil {
		return c.JSON(400, bson.M{"error": "invalid session"})
	}

	user, err := convertDatabaseUser(&session.User)

	if err != nil {
		graviton.Logger.Warn("Could not look up roles for user", zap.String("Email", user.Email), zap.Error(err))
		return c.JSON(502, bson.M{"error": "database lookup error"})
	}

	return c.JSON(200, user)
}

func Logout(c echo.Context) error {
	token := extractBearer(c)
	err := deleteSession(token)

	if err != nil {
		graviton.Logger.Warn("Error deleting session", zap.String("Token", token), zap.Error(err))
		return c.JSON(502, bson.M{"error": "unable to delete session"})
	}

	graviton.Logger.Info("Session ended", zap.String("Token", token))
	return c.JSON(200, bson.M{"status": "ok"})
}

func getCookie(token string) *http.Cookie {
	return &http.Cookie{
		Name:    "graviton_bearer",
		Value:   token,
		Expires: time.Now().Add(24 * time.Hour),
		//Secure: true,
		//HttpOnly: true,
		Path: "/",
	}
}
