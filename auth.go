package jobsearch

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/google/go-github/github"
	"golang.org/x/oauth2"
)

// TODO: change this!!!!
var signingKey = []byte("Secret signing key")

// AuthJWT contains the JWT AuthToken to be used for future authorization.
type AuthJWT struct {
	BearerToken string `json:"bearer_token" bson:"bearer_token"`
}

// AuthToken is the token used for authentication into the application, that is
// inspected by the MustAuth middleware on routes requiring authentication.
type AuthToken struct {
	Login   string    `json:"login" bson:"login"`
	Expires time.Time `json:"exp" bson:"exp"`
}

// NewAuthToken generates a new auth token struct containing the user's username
// and the expiration time.
func NewAuthToken(login string) *AuthToken {
	at := &AuthToken{
		Login:   login,
		Expires: time.Now().Add(time.Hour * 24),
	}
	return at
}

// Tokenize converts a State instance into a signed JWT.
func (s AuthToken) Tokenize() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"login": s.Login,
		"exp":   s.Expires,
	})

	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ParseAuthToken creates a AuthToken instance from a signed JWT that was formed from a
// previous AuthToken instance.
func ParseAuthToken(token string) (*AuthToken, error) {
	t, err := jwt.Parse(token, func(*jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims := t.Claims.(jwt.MapClaims)

	exp, err := time.Parse(time.RFC3339, claims["exp"].(string))
	if err != nil {
		return nil, err
	}
	at := &AuthToken{
		Login:   claims["login"].(string),
		Expires: exp,
	}
	log.Println("AuthToken login:", at.Login)
	return at, nil
}

// State consists of the state of the current request and a a randomly generated
// string, for use in OAuth requests.
type State struct {
	RequestURL   string    `json:"request_url" bson:"request_url"`
	RandomString string    `json:"random_string" bson:"random_string"`
	Expires      time.Time `json:"exp" bson:"exp"`
}

// NewState generates a new state struct containing the request url and a random string.
func NewState(url string) (*State, error) {
	p := make([]byte, 32)
	if _, err := rand.Read(p); err != nil {
		return nil, err
	}
	s := &State{
		RequestURL:   url,
		RandomString: base64.StdEncoding.EncodeToString(p),
		Expires:      time.Now().Add(time.Hour * 24),
	}
	return s, nil
}

// Tokenize converts a State instance into a signed JWT.
func (s State) Tokenize() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"request_url":   s.RequestURL,
		"random_string": s.RandomString,
		"exp":           s.Expires,
	})

	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

// ParseState creates a state instance from a signed JWT that was formed from a
// state instance.
func ParseState(token string) (*State, error) {
	t, err := jwt.Parse(token, func(*jwt.Token) (interface{}, error) {
		return signingKey, nil
	})
	if err != nil {
		return nil, err
	}
	claims := t.Claims.(jwt.MapClaims)

	exp, err := time.Parse(time.RFC3339, claims["exp"].(string))
	if err != nil {
		return nil, err
	}
	state := &State{
		RequestURL:   claims["request_url"].(string),
		RandomString: claims["random_string"].(string),
		Expires:      exp,
	}
	return state, nil
}

// AuthController has methods that serve as handlers for the authentication process.
type AuthController struct {
	*oauth2.Config
	*DB
}

// Login begins the login process using OAuth 2.0 from a Github account.
func (a AuthController) Login(w http.ResponseWriter, r *http.Request) {
	//TODO: capture state of request
	state, err := NewState("")
	if err != nil {
		http.Error(w, "Error creating login token", http.StatusInternalServerError)
		return
	}

	log.Println("Login state random string created:", state.RandomString)

	if err = a.DB.SaveAuthState(state); err != nil {
		http.Error(w, "Error saving login state", http.StatusInternalServerError)
		return
	}

	token, err := state.Tokenize()
	if err != nil {
		http.Error(w, "Error creating login token", http.StatusInternalServerError)
		return
	}

	urlStr := a.Config.AuthCodeURL(token)
	http.Redirect(w, r, urlStr, http.StatusTemporaryRedirect)
}

// Callback completes the Github OAuth 2.0 authentication process.
// If successful
func (a AuthController) Callback(w http.ResponseWriter, r *http.Request) {
	t := r.FormValue("state")

	state, err := ParseState(t)
	if err != nil {
		http.Error(w, "Failed to parse state token", http.StatusInternalServerError)
		return
	}

	if state.Expires.Before(time.Now()) {
		http.Error(w, "Authentication token has expired:"+state.Expires.String(), http.StatusUnauthorized)
		return
	}
	log.Println("Login state random string received:", state.RandomString)
	valid, err := a.DB.IsValidAuthState(state.RandomString)
	if !valid || err != nil {
		log.Printf("invalid oauth state %s", state.RandomString)
		http.Error(w, "Invalid OAuth state token", http.StatusUnauthorized)
		return
	}
	code := r.FormValue("code")
	token, err := a.Config.Exchange(context.Background(), code)
	if err != nil {
		log.Printf("oauthConf.Exchange() failed with '%s'\n", err)
		http.Error(w, "Failed to retrieve GitHub authentication", http.StatusInternalServerError)
		return
	}

	oauthClient := a.Config.Client(oauth2.NoContext, token)
	client := github.NewClient(oauthClient)
	gu, _, err := client.Users.Get(context.Background(), "")
	user := NewUserFromGithub(gu)
	if err != nil {
		log.Printf("client.Users.Get() faled with '%s'\n", err)
		http.Error(w, "Failed to retrieve GitHub authentication", http.StatusInternalServerError)
		return
	}
	log.Printf("Logged in as GitHub user: %s\n", user.Login)
	at, err := NewAuthToken(user.Login).Tokenize()
	if err != nil {
		log.Printf("AuthToken.Tokenize() failed with '%s'\n", err)
		http.Error(w, "Failed to generate auth token", http.StatusInternalServerError)
		return
	}
	if err := a.DB.UpsertUser(user); err != nil {
		log.Printf("db.UpsertUser failed with '%s'\n", err)
		http.Error(w, "Failed to save user", http.StatusInternalServerError)
		return
	}
	response := &AuthJWT{BearerToken: at}
	if err := a.DB.SaveAuthJWT(response); err != nil {
		log.Printf("DB.SaveAuthJWT() failed with '%s'\n", err)
		http.Error(w, "Failed to save auth token", http.StatusInternalServerError)
		return
	}
	respond(w, r, http.StatusOK, response)
}

// Logout logs the user out of the application.
func (a AuthController) Logout(w http.ResponseWriter, r *http.Request) {
	jwt, err := extractJWT(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	a.DB.RemoveAuthJWT(jwt)
}

func extractJWT(r *http.Request) (string, error) {
	head := strings.Split(r.Header.Get("Authorization"), " ")
	if head[0] != "Bearer" {
		return "", errors.New("Authorization header not set properly with Bearer")
	}
	return head[1], nil
}
