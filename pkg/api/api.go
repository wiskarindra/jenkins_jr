package api

// API contains informations needed for an API
type API struct {
	Method    string
	Endpoint  string
	Action    string
	Authority Authority
	Handle    HandleWithError
}

// Authority represents authority of users
type Authority int

const (
	// Anonymous represents users not logged in
	Anonymous Authority = iota
	// User represents normal users
	User
	// Admin represents trusted users (admins)
	Admin
)

// StartAPIs starts API handlers
func StartAPIs(router Router, apis []API) {
	for _, api := range apis {
		router.HandlerFunc(api.Method, api.Endpoint, middleware(api.Action, api.Authority, api.Handle))
	}
}
