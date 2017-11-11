package auth

import (
	"strings"
	"time"

	"github.com/jslater89/graviton"
	"github.com/labstack/echo"
	"go.uber.org/zap"
	"gopkg.in/mgo.v2/bson"
)

func extractBearer(c echo.Context) string {
	authHeader := c.Request().Header.Get("Authorization")
	splitHeader := strings.Split(authHeader, " ")

	if len(splitHeader) == 2 {
		return splitHeader[1]
	}

	cookie, err := c.Cookie("graviton_bearer")

	if err == nil {
		return cookie.Value
	}
	return ""
}

func checkSessionExpiration(session *Session) bool {
	if time.Now().After(session.ExpiresAt) {
		return false
	}
	return true
}

func IsAuthorized(c echo.Context, path string) bool {
	bearer := extractBearer(c)
	sess, err := getSession(bearer)

	if err != nil {
		graviton.Logger.Info("Error getting session", zap.Error(err))
		c.JSON(401, bson.M{"error": "not logged in"})
		return false
	}

	if !checkSessionExpiration(sess) {
		graviton.Logger.Info("Session expired")
		deleteSession(bearer)
		c.JSON(401, bson.M{"error": "login expired"})
	}

	// TODO: user authorized for path? Check role permissions for exact match;
	// if not, use rule with the longest prefix
	// canRead == GET, canWrite == everything else
	// c.JSON(403, bson.M{"error": "not authorized for resource"})

	user := sess.User
	graviton.Logger.Info("User authorized for path", zap.String("User", user.Email), zap.String("Path", path))
	sess.ExpiresAt = time.Now().Add(1 * time.Hour)
	saveSession(*sess)

	return true
}
