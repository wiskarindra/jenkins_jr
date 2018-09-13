package api

import (
	"context"
	"net/http"
	"time"

	"github.com/bukalapak/jenkins_jr/pkg/currentuser"
	"github.com/bukalapak/jenkins_jr/pkg/log"
	"github.com/bukalapak/jenkins_jr/pkg/resource"

	"github.com/bukalapak/apinizer/response"
	"github.com/bukalapak/packen/instrument"
	"github.com/google/uuid"
)

// HandleWithError is http.HandlerFunc that return an error
type HandleWithError func(http.ResponseWriter, *http.Request) error

func middleware(action string, security Authority, handle HandleWithError) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
		defer cancel()

		startTime := time.Now()

		reqID, err := createRequestID(r)
		ctx = resource.NewContext(ctx, reqID, action, startTime)
		if err != nil {
			log.ErrLog(ctx, err, "request-ID", "generating request-ID fail")
		}
		r.Header.Set("X-Request-ID", reqID)

		currentUser, err := currentuser.FromRequest(r)
		if err != nil {
			log.ErrLog(ctx, err, "authorization", "authorize fail")
			response.Write(w, response.BuildError(InvalidTokenError), InvalidTokenError.HTTPCode)
			return
		}

		if !isRequestAuthorized(security, currentUser) {
			response.Write(w, response.BuildError(UserUnauthorizedError), UserUnauthorizedError.HTTPCode)
			return
		}

		ctx = currentuser.NewContext(ctx, currentUser)

		status := "ok"
		err = handle(w, r.WithContext(ctx))
		if err != nil {
			status = "fail"
		}

		instrument.ObserveLatency(r.Method, action, status, time.Since(startTime).Seconds())
	}
}

func isRequestAuthorized(security Authority, currentUser *currentuser.CurrentUser) bool {
	switch security {
	case Admin:
		return isRoleAllowed(currentUser.Role)
	case User:
		return isUserLoggedIn(currentUser.ID)
	default:
		return true
	}
}

func isRoleAllowed(role string) bool {
	return isInSliceString(role, []string{"admin", "marketing", "sales", "ultraman"})
}

func isUserLoggedIn(userID uint) bool {
	return userID != 0
}

func createRequestID(r *http.Request) (string, error) {
	reqID := r.Header.Get("X-Request-ID")
	if reqID == "" {
		temp, err := uuid.NewRandom()
		if err != nil {
			return "", err
		}
		return temp.String(), nil
	}
	return reqID, nil
}
