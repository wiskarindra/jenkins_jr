package log

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/wiskarindra/jenkins_jr/pkg/currentuser"
	"github.com/wiskarindra/jenkins_jr/pkg/resource"
)

// DevLog logs only on development or staging
func DevLog(v ...interface{}) {
	if os.Getenv("ENV") == "development" || os.Getenv("ENV") == "staging" {
		log.Println(v)
	}
}

// ErrLog logs errors with packen.RequestError
func ErrLog(ctx context.Context, err error, category, message string) {
	if os.Getenv("ENV") != "test" {
		res := resource.FromContext(ctx)
		plog.RequestError(fmt.Sprintf("%s", err.Error()),
			plog.NewField("request_id", res.RequestID),
			plog.NewField("tags", append([]string{"post"}, res.Action, category)),
			plog.NewField("message", message),
			plog.NewField("duration", strconv.FormatFloat(time.Since(res.StartTime).Seconds(), 'f', -1, 64)),
		)
	}
}

// InfoLog logs informations with packen.RequestInfo
func InfoLog(ctx context.Context, message string, tags ...string) {
	if os.Getenv("ENV") != "test" {
		res := resource.FromContext(ctx)
		user := currentuser.FromContext(ctx)
		plog.RequestInfo("",
			plog.NewField("request_id", res.RequestID),
			plog.NewField("tags", append([]string{"post"}, tags...)),
			plog.NewField("message", fmt.Sprintf("%d: %s", user.ID, message)),
			plog.NewField("duration", strconv.FormatFloat(time.Since(res.StartTime).Seconds(), 'f', -1, 64)),
		)
	}
}

func Fatal(v ...interface{}) {
	log.Fatal(v)
}
