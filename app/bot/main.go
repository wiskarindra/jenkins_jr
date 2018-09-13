package main

import (
	"net/http"

	"github.com/wiskarindra/jenkins_jr/pkg/api"
	"github.com/wiskarindra/jenkins_jr/pkg/log"
	"github.com/wiskarindra/jenkins_jr/pkg/mysql"
	"github.com/wiskarindra/jenkins_jr/pkg/jenkins_jr"

	"github.com/rs/cors"
	"github.com/subosito/gotenv"
)

func main() {
	gotenv.Load()

	db := mysql.Init()
	env := jenkins_jr.Env{DB: db}

	router := api.NewRouter()
	router.HandlerFunc("GET", "/metrics", instrument.Handler)
	router.HandlerFunc("GET", "/healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	co := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
		AllowedMethods: []string{"GET", "POST", "PATCH", "DELETE", "PUT", "HEAD", "OPTIONS"},
		AllowedHeaders: []string{"*"},
		MaxAge:         86400,
	})
	log.DevLog("Listening at port 666")
	log.Fatal(http.ListenAndServe(":666", co.Handler(router)))
}
