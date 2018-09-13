package api

import (
	"net/http"

	gorilla "github.com/gorilla/mux"
)

type Router interface {
	HandlerFunc(method, path string, handler http.HandlerFunc)
	ServeHTTP(http.ResponseWriter, *http.Request)
}

type mux struct {
	r *gorilla.Router
}

func NewRouter() Router {
	r := gorilla.NewRouter()
	return &mux{r: r}
}

func (m *mux) HandlerFunc(method, path string, handler http.HandlerFunc) {
	m.r.HandleFunc(path, handler).Methods(method)
}

func (m *mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	m.r.ServeHTTP(w, req)
}

type Params map[string]string

func (ps Params) ByName(name string) string {
	return ps[name]
}

func GetParams(r *http.Request) Params {
	return gorilla.Vars(r)
}

func SetParams(r *http.Request, val map[string]string) *http.Request {
	return gorilla.SetURLVars(r, val)
}
