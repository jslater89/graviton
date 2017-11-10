package auth

import (
	"fmt"
	"math"
	"net/http"
	"os"

	"github.com/gorilla/sessions"
	"github.com/jslater89/graviton/config"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/gplus"
)

func InitOauth() {
	goth.UseProviders(gplus.New(config.GetConfig().GoogleClientID, config.GetConfig().GoogleSecret, "http://localhost:10000/api/v1/auth/google/callback"))
	gothic.GetProviderName = func(*http.Request) (string, error) {
		return "gplus", nil
	}
	store := sessions.NewFilesystemStore(os.TempDir(), []byte("goth-example"))

	// set the maxLength of the cookies stored on the disk to a larger number to prevent issues with:
	// securecookie: the value is too long
	// when using OpenID Connect , since this can contain a large amount of extra information in the id_token

	// Note, when using the FilesystemStore only the session.ID is written to a browser cookie, so this is explicit for the storage on disk
	store.MaxLength(math.MaxInt64)

	gothic.Store = store
}

func HandleUser(user goth.User) {
	fmt.Println(user)
}
