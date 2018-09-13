package api

import (
	"net/http"

	"github.com/bukalapak/apinizer/response"
)

// IndexMeta contains metadata of an index response
type IndexMeta struct {
	HTTPStatus int    `json:"http_status"`
	Limit      uint64 `json:"limit"`
	Offset     uint64 `json:"offset"`
	Total      int    `json:"total"`
	TotalPages uint   `json:"total_pages,omitempty"`
}

var (
	// InvalidParameterError represents Invalid parameter error
	InvalidParameterError = response.CustomError{
		Message:  "Invalid parameter",
		Code:     70111,
		HTTPCode: http.StatusUnprocessableEntity,
	}
	// PostNotFoundError represents Post not found error
	PostNotFoundError = response.CustomError{
		Message:  "Post not found",
		Code:     70112,
		HTTPCode: http.StatusNotFound,
	}
	// InvalidTokenError represents Invalid token error
	InvalidTokenError = response.CustomError{
		Message:  "Invalid token",
		Code:     70113,
		HTTPCode: http.StatusUnauthorized,
	}
	// UserUnauthorizedError represents User unauthorized error
	UserUnauthorizedError = response.CustomError{
		Message:  "User unauthorized",
		Code:     70114,
		HTTPCode: http.StatusForbidden,
	}
	// InfluencerNotFoundError represents Influencer not found error
	InfluencerNotFoundError = response.CustomError{
		Message:  "Influencer not found",
		Code:     70115,
		HTTPCode: http.StatusNotFound,
	}
)
