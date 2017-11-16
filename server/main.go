package main

import (
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

	s, err := mgo.Dial("localhost")
	if err != nil {
		log.Fatalln("Failed to dial ")
	}
	db := jobsearch.NewDB(s)

	auth := jobsearch.AuthController{
		Config: oauthConfig,
		DB:     db,
	}
	uc := jobsearch.NewUserController(db)

	authMiddleware := jobsearch.NewMiddleware(uc.MustAuth)

	r.HandleFunc("/", notImplemented)
	r.HandleFunc("/login", auth.Login).Methods(http.MethodGet)
	r.HandleFunc("/callback", auth.Callback)
	r.HandleFunc("/logout", auth.Logout)

	r.Handle("/jobs", authMiddleware.AdaptFunc(uc.GetAllJobs)).Methods(http.MethodGet)
	r.Handle("/jobs", authMiddleware.AdaptFunc(uc.PostJob)).Methods(http.MethodPost)
	r.Handle("/jobs/{id}", authMiddleware.AdaptFunc(uc.GetJobByID)).Methods(http.MethodGet)
	r.HandleFunc("/jobs/{id}", notImplemented).Methods(http.MethodDelete)
	r.HandleFunc("/jobs/{id}/apply", notImplemented).Methods(http.MethodPut)

	r.Handle("/user", uc.MustAuth(http.HandlerFunc(uc.GetUser)))

	log.Println("Listening on port :9090")
	http.ListenAndServe(":9090", r)
}

var notImplemented = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "NOT IMPLEMENTED YET", http.StatusNotFound)
	})
