package resource

import (
	"context"
	"time"
)

// Resources contains request resources
type Resources struct {
	RequestID string
	Action    string
	StartTime time.Time
}

type key int

// Key is resources context key
const Key key = 0

// NewContext returns context containing given CurrentUser
func NewContext(ctx context.Context, rID string, action string, startTime time.Time) context.Context {
	return context.WithValue(ctx, Key, &Resources{
		RequestID: rID,
		Action:    action,
		StartTime: startTime,
	})
}

// FromContext returns Resources contained in given context
func FromContext(ctx context.Context) *Resources {
	res, _ := ctx.Value(Key).(*Resources)
	return res
}
