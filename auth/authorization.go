package auth

import (
	"encoding/hex"
	"net/http"
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

// IsAuthorized checks the session included in the request. If
// the user is not authorized for the given request, returns false and
// makes an appropriate response with the context. Otherwise, returns
// true and defers responses to the caller. If the user is authorized
// for a given endpoint, IsAuthorized extends the session's expiration.
func IsAuthorized(c echo.Context, path string) bool {
	bearer := extractBearer(c)

	if bearer == getOrCreateAPIKey(false) {
		return true
	}

	sess, err := getSession(bearer)

	if err != nil {
		graviton.Logger.Info("Error getting session", zap.String("Bearer", bearer), zap.Error(err))
		c.JSON(401, bson.M{"error": "not logged in"})
		return false
	}

	if !checkSessionExpiration(sess) {
		graviton.Logger.Info("Session expired for user", zap.String("Email", sess.User.Email))
		deleteSession(bearer)
		c.JSON(401, bson.M{"error": "login expired"})
		return false
	}

	user := &sess.User

	writeRequest := (c.Request().Method != http.MethodGet)
	if !checkUserPermissions(user, writeRequest, path) {
		graviton.Logger.Info("User not authorized for resource", zap.String("Email", user.Email), zap.String("Path", path))
		c.JSON(403, bson.M{"error": "not authorized for resource"})
		return false
	}

	graviton.Logger.Info("User authorized for path", zap.String("User", user.Email), zap.String("Path", path))
	sess.ExpiresAt = time.Now().Add(1 * time.Hour)
	saveSession(*sess)

	return true
}

func GetAPIKey(c echo.Context) error {
	if !IsAuthorized(c, "/auth/apikey") {
		return nil
	}

	return c.JSON(200, bson.M{"key": getOrCreateAPIKey(false)})
}

func ResetAPIKey(c echo.Context) error {
	if !IsAuthorized(c, "/auth/apikey") {
		return nil
	}

	return c.JSON(200, bson.M{"key": getOrCreateAPIKey(true)})
}

type apiDoc struct {
	ID  bson.ObjectId `bson:"_id,omitempty"`
	Key string        `bson:"key"`
}

func getOrCreateAPIKey(reset bool) string {
	n, err := db.apiKeyCollection.Find(bson.M{}).Count()

	if err != nil {
		panic(err)
	}

	if n == 0 || reset {
		db.apiKeyCollection.RemoveAll(bson.M{})

		keyBytes := []byte(generateSessionToken() + generateSessionToken() + generateSessionToken())
		keyString := hex.EncodeToString(keyBytes)

		key := apiDoc{
			ID:  bson.NewObjectId(),
			Key: keyString,
		}

		db.apiKeyCollection.Insert(key)
	}

	key := &apiDoc{}
	db.apiKeyCollection.Find(bson.M{}).One(key)

	return key.Key
}

func checkUserPermissions(user *User, write bool, path string) bool {
	roles, err := getUserRoles(user)

	if err != nil {
		graviton.Logger.Warn("User role lookup error", zap.String("Email", user.Email), zap.Error(err))
		return false
	}

	permissions := []Permission{}
	for _, role := range roles {
		permissions = append(permissions, role.Permissions...)
	}

	bestMatchLength := 0
	var bestPermission Permission

	for _, permission := range permissions {
		if strings.HasPrefix(path, permission.Path) && len(permission.Path) > bestMatchLength {
			bestMatchLength = len(permission.Path)
			bestPermission = permission

			// Stop searching on exact match
			if bestPermission.Path == path {
				break
			}
		}
	}

	return write && bestPermission.CanWrite || !write && bestPermission.CanRead
}

func verifyBaseRoles() error {
	n, err := db.roleCollection.Find(bson.M{"name": "Viewer"}).Count()

	if err == nil && n > 0 {
		return nil
	}
	graviton.Logger.Info("Making default roles")

	viewerRole := Role{
		ID:   bson.NewObjectId(),
		Name: "Viewer",
		Permissions: []Permission{
			Permission{
				Path:     "/",
				CanRead:  true,
				CanWrite: false,
			},
			Permission{
				Path:     "/auth/apikey",
				CanRead:  false,
				CanWrite: false,
			},
		},
	}

	db.roleCollection.UpsertId(viewerRole.ID, viewerRole)

	editorRole := Role{
		ID:   bson.NewObjectId(),
		Name: "Editor",
		Permissions: []Permission{
			Permission{
				CanRead:  true,
				CanWrite: true,
				Path:     "/",
			},
		},
	}

	db.roleCollection.UpsertId(editorRole.ID, editorRole)
	return nil
}
