package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dkundathagard/jobsearch"
	"github.com/gorilla/mux"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	mgo "gopkg.in/mgo.v2"
)

func main() {
	r := mux.NewRouter()

	oauthConfig := &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		Endpoint:     github.Endpoint,
		RedirectURL:  "http://localhost:9090/callback",
	}

	db, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatalln("Failed to dial ")
	}

	auth := jobsearch.AuthController{
		Config: oauthConfig,
		DB:     jobsearch.NewDB(db),
	}

	r.HandleFunc("/", notImplemented)
	r.HandleFunc("/login", auth.Login).Methods(http.MethodGet)
	r.HandleFunc("/callback", auth.Callback)
	r.Handle("crap", auth.MustAuth(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		email := w.Header().Get("user_email")
		fmt.Fprintf(w, email)
	})))

	r.HandleFunc("/jobs", notImplemented).Methods(http.MethodGet)
	r.HandleFunc("/jobs", notImplemented).Methods(http.MethodPost)
	r.HandleFunc("/jobs/{id}", notImplemented).Methods(http.MethodGet)
	r.HandleFunc("/jobs/{id}", notImplemented).Methods(http.MethodDelete)
	r.HandleFunc("/jobs/{id}/apply", notImplemented).Methods(http.MethodPut)

	log.Println("Listening on port :9090")
	http.ListenAndServe(":9090", r)
}

var notImplemented = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "NOT IMPLEMENTED YET", http.StatusNotFound)
	})
