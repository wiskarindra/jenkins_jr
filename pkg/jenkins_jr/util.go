package jenkins_jr

import (
	"context"
	"fmt"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/wiskarindra/jenkins_jr/config"
	"github.com/wiskarindra/jenkins_jr/pkg/log"
)

func getLimitOffsetFromURLQuery(uq url.Values) (uint64, uint64) {
	limit, err := strconv.ParseUint(uq.Get("limit"), 10, 8)
	if err != nil || limit > 25 || limit == 0 {
		limit = 10
	}

	offset, err := strconv.ParseUint(uq.Get("offset"), 10, 8)
	if err != nil || offset > 10000 {
		offset = 0
	}

	return limit, offset
}

func isInSliceInt64(v int64, slice []int64) bool {
	for _, s := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func isInSliceString(v string, slice []string) bool {
	for _, s := range slice {
		if v == s {
			return true
		}
	}
	return false
}

func isUserLoggedIn(userID uint) bool {
	return userID != 0
}

func joinSliceInt64(values []int64) string {
	strs := make([]string, 0)
	for _, v := range values {
		strs = append(strs, fmt.Sprint(v))
	}
	return strings.Join(strs, ",")
}

func parseIntContext(ctx context.Context, s string) (int64, error) {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.ErrLog(ctx, err, "parse-params", "parse int64 fail")
	}
	return i, err
}

func reflectValuesToString(values []reflect.Value) []string {
	var temp []string
	for _, v := range values {
		temp = append(temp, v.Interface().(string))
	}
	return temp
}

func timeNowDB() string {
	t := time.Now().UTC()
	return timeToStringDB(&t)
}

func timeToStringDB(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(config.DatabaseDatetimeFormat)
}

func timeNowResponse() string {
	t := time.Now().UTC()
	return timeToStringResponse(&t)
}

func timeToStringResponse(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(config.ResponseDatetimeFormat)
}
