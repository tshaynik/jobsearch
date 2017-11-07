package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	r := mux.NewRouter()

	r.HandleFunc("/", notImplemented)
	r.HandleFunc("/login", notImplemented)
	r.HandleFunc("/callback", notImplemented)

	r.HandleFunc("/jobs", notImplemented).Methods(http.MethodGet)
	r.HandleFunc("/jobs", notImplemented).Methods(http.MethodPost)
	r.HandleFunc("/jobs/{id}", notImplemented).Methods(http.MethodGet)
	r.HandleFunc("/jobs/{id}", notImplemented).Methods(http.MethodDelete)
	r.HandleFunc("/jobs/{id}/apply", notImplemented).Methods(http.MethodPut)

	http.ListenAndServe(":8080", r)
}

var notImplemented = http.HandlerFunc(
	func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	})
