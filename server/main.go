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

type config struct {
	Port               string
	DBConfig           string
	GithubClientID     string
	GithubClientSecret string
}

func newConfig() *config {
	port := os.Getenv("JOBSEARCH_PORT")
	if port == "" {
		port = ":8080"
	}
	dbconfig := os.Getenv("JOBSEARCH_DB")
	if dbconfig == "" {
		dbconfig = "localhost"
	}
	ghID := os.Getenv("GITHUB_CLIENT_ID")
	if ghID == "" {
		log.Fatalln("GITHUB_CLIENT_ID environtment variable must be set to enable login.")
	}
	ghSecret := os.Getenv("GITHUB_CLIENT_SECRET")
	if ghSecret == "" {
		log.Fatalln("GITHUB_CLIENT_SECRET environtment variable must be set to enable login.")
	}
	return &config{
		Port:               port,
		DBConfig:           dbconfig,
		GithubClientID:     ghID,
		GithubClientSecret: ghSecret,
	}
}

func main() {
	r := mux.NewRouter()

	config := newConfig()

	oauthConfig := &oauth2.Config{
		ClientID:     config.GithubClientID,
		ClientSecret: config.GithubClientSecret,
		Endpoint:     github.Endpoint,
		RedirectURL:  "http://localhost:9090/callback",
	}

	s, err := mgo.Dial(config.DBConfig)
	if err != nil {
		log.Fatalln("Failed to dial ")
	}
	db := jobsearch.NewDB(s)

	auth := jobsearch.AuthController{
		Config: oauthConfig,
		DB:     db,
	}
	uc := jobsearch.NewUserController(db)

	authMiddleware := jobsearch.NewMiddleware(
		jobsearch.LogRequest(log.New(os.Stdout, "\nrequest log | ", log.LstdFlags)),
		uc.MustAuth,
	)

	r.HandleFunc("/", notImplemented)
	r.HandleFunc("/login", auth.Login).Methods(http.MethodGet)
	r.HandleFunc("/callback", auth.Callback).Methods(http.MethodGet)
	r.HandleFunc("/logout", auth.Logout)

	r.Handle("/jobs", authMiddleware.AdaptFunc(uc.GetAllJobs)).Methods(http.MethodGet)
	r.Handle("/jobs", authMiddleware.AdaptFunc(uc.PostJob)).Methods(http.MethodPost)
	r.Handle("/jobs/{id}", authMiddleware.AdaptFunc(uc.GetJobByID)).Methods(http.MethodGet)
	r.Handle("/jobs/{id}", authMiddleware.AdaptFunc(uc.DeleteJobByID)).Methods(http.MethodDelete)
	r.HandleFunc("/jobs/{id}/apply", notImplemented).Methods(http.MethodPut)

	r.Handle("/user", uc.MustAuth(http.HandlerFunc(uc.GetUser)))

	log.Println("Listening on port", config.Port)
	http.ListenAndServe(config.Port, r)
}

var notImplemented = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "NOT IMPLEMENTED YET", http.StatusNotFound)
	})
