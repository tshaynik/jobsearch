package jobsearch

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/google/go-github/github"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
)

// User is an authenticated user of the jobsearch app.
type User struct {
	ID        bson.ObjectId `json:"id" bson:"_id,omitempty"`
	Login     string        `json:"login" bson:"login"`
	AvatarURL string        `json:"avatar_url" bson:"avatar_url"`
}

// NewUserFromGithub creates a new user instance from a github user.
func NewUserFromGithub(gu *github.User) *User {
	return &User{
		Login:     gu.GetLogin(),
		AvatarURL: gu.GetAvatarURL(),
	}
}

// UserController provides dependency injection for user requests.
type UserController struct {
	*DB
}

type contextKey string

// NewUserController returns a new user controller with the specified database
// pointer.
func NewUserController(db *DB) *UserController {
	return &UserController{DB: db}
}

// MustAuth is HTTP middleware for the authorization of a resource, which validates
// the JWT passed in the Authorization header, and if authorized, passes the user
// information to the request context.
func (uc UserController) MustAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		jwt, err := extractJWT(r)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		if !uc.DB.ValidateAuthJWT(jwt) {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		at, err := ParseAuthToken(jwt)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		ctx := context.WithValue(r.Context(), contextKey("login"), at.Login)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	})
}

// GetUser handler func returns information about the authenticated user.
func (uc UserController) GetUser(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(contextKey("login")).(string)
	user, err := uc.DB.GetUser(login)
	if err != nil {
		http.Error(w, "Failed to get user info.", http.StatusInternalServerError)
		return
	}
	respond(w, r, http.StatusOK, user)
}

// PostJob adds a new job to the database.
func (uc UserController) PostJob(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(contextKey("login")).(string)
	var j Job
	if err := json.NewDecoder(r.Body).Decode(&j); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	j.ID = bson.NewObjectId()
	j.UserLogin = login
	if err := uc.DB.CreateJob(&j); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// GetAllJobs adds a new job to the database.
func (uc UserController) GetAllJobs(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(contextKey("login")).(string)
	jobs, err := uc.DB.FindAllJobs(login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, r, http.StatusOK, jobs)
}

// GetJobByID adds a new job to the database.
func (uc UserController) GetJobByID(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(contextKey("login")).(string)
	id := mux.Vars(r)["id"]
	job, err := uc.DB.FindJobByID(login, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, r, http.StatusOK, job)
}

// DeleteJobByID adds a new job to the database.
func (uc UserController) DeleteJobByID(w http.ResponseWriter, r *http.Request) {
	login := r.Context().Value(contextKey("login")).(string)
	id := mux.Vars(r)["id"]
	if err := uc.DB.RemoveJob(login, id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	respond(w, r, http.StatusOK, nil)
}
