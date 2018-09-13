// Package response is used to write response based on Bukalapak API V4 standard
package response

import (
	"encoding/json"
	"net/http"
)

var (
	// ErrTetapTenangTetapSemangat custom error on unexpected error
	ErrTetapTenangTetapSemangat = CustomError{
		Message:  "Tetap Tenang Tetap Semangat",
		Code:     999,
		HTTPCode: http.StatusInternalServerError,
	}
)

// SuccessBody holds data for success response
type SuccessBody struct {
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Meta    interface{} `json:"meta"`
}

// ErrorBody holds data for error response
type ErrorBody struct {
	Errors []ErrorInfo `json:"errors"`
	Meta   interface{} `json:"meta"`
}

// MetaInfo holds meta data
type MetaInfo struct {
	HTTPStatus int `json:"http_status"`
}

// ErrorInfo holds error detail
type ErrorInfo struct {
	Message string `json:"message"`
	Code    int    `json:"code"`
	Field   string `json:"field,omitempty"`
}

// CustomError holds data for customized error
type CustomError struct {
	Message  string
	Field    string
	Code     int
	HTTPCode int
}

// Error is a function to convert error to string.
// It exists to satisfy error interface
func (c CustomError) Error() string {
	return c.Message
}

// BuildSuccess is a function to create SuccessBody
func BuildSuccess(data interface{}, message string, meta interface{}) SuccessBody {
	return SuccessBody{
		Data:    data,
		Message: message,
		Meta:    meta,
	}
}

// BuildError is a function to create ErrorBody
func BuildError(errors ...error) ErrorBody {
	var (
		ce CustomError
		ok bool
	)

	if len(errors) == 0 {
		ce = ErrTetapTenangTetapSemangat
	} else {
		err := errors[0]
		ce, ok = err.(CustomError)
		if !ok {
			ce = ErrTetapTenangTetapSemangat
		}
	}

	return ErrorBody{
		Errors: []ErrorInfo{
			{
				Message: ce.Message,
				Code:    ce.Code,
				Field:   ce.Field,
			},
		},
		Meta: MetaInfo{
			HTTPStatus: ce.HTTPCode,
		},
	}
}

// Write is a function to write data in json format
func Write(w http.ResponseWriter, result interface{}, status int) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(result)
}
