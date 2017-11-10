package api

import (
	"fmt"

	"github.com/jslater89/graviton/auth"
	"github.com/labstack/echo"
	"github.com/markbates/goth/gothic"
	"gopkg.in/mgo.v2/bson"
)

func GoogleAuthLogin(c echo.Context) error {
	// try to get the user without re-authenticating
	if user, err := gothic.CompleteUserAuth(c.Response(), c.Request()); err == nil {
		auth.HandleUser(user)
	} else {
		gothic.BeginAuthHandler(c.Response(), c.Request())
	}

	return c.JSON(200, bson.M{"status": "ok"})
}

func GoogleAuthCallback(c echo.Context) error {
	user, err := gothic.CompleteUserAuth(c.Response(), c.Request())
	if err != nil {
		fmt.Println(err)
		return c.JSON(502, bson.M{"error": err.Error()})
	}

	auth.HandleUser(user)
	return c.JSON(200, bson.M{"status": "ok"})
}
