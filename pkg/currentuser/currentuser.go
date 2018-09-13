package currentuser

import (
	"context"
	"net/http"
	"os"
	"strconv"

	"github.com/bukalapak/apinizer/authorization"
)

// CurrentUser contains user informations in a context
type CurrentUser struct {
	ID   uint
	Role string

	applicationID int
	appVersion    string
}

type key int

// Key is currentuser context key
const Key key = 0

// FromRequest returns CurrentUser contained in given request
func FromRequest(r *http.Request) (*CurrentUser, error) {
	token, err := authorization.NewJWTAuthorization(r)
	if err != nil {
		return nil, err
	}

	return &CurrentUser{
		ID:   token.Token.ResourceOwner.ID,
		Role: token.Token.ResourceOwner.Role,

		applicationID: token.Token.ApplicationID,
		appVersion:    r.Header.Get("Bukalapak-App-Version"),
	}, nil
}

// NewContext returns context containing given CurrentUser
func NewContext(ctx context.Context, user *CurrentUser) context.Context {
	return context.WithValue(ctx, Key, user)
}

// FromContext returns CurrentUser contained in given context
func FromContext(ctx context.Context) *CurrentUser {
	user, _ := ctx.Value(Key).(*CurrentUser)
	return user
}

// Platform returns given user's platform type
func (user *CurrentUser) Platform() string {
	switch strconv.Itoa(user.applicationID) {
	case os.Getenv("BUKALAPAK_ANDROID_APP_ID"):
		return "blandroid"
	case os.Getenv("BUKALAPAK_IOS_APP_ID"):
		return "blios"
	}
	return ""
}

// AppVersion returns given user's platform version
func (user *CurrentUser) AppVersion() int {
	av, _ := strconv.Atoi(user.appVersion)
	return av
}
