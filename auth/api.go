package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/jslater89/graviton"

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
		fmt.Println(err)
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
	return nil
}

func getCookie(token string) *http.Cookie {
	return &http.Cookie{
		Domain:  "localhost:10000",
		Name:    "graviton_bearer",
		Value:   token,
		Expires: time.Now().Add(24 * time.Hour),
		//Secure: true,
		//HttpOnly: true,
		Path: "/",
	}
}
