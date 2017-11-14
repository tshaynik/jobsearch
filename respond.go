package jobsearch

import (
	"encoding/json"
	"net/http"
)

func respond(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		panic("Could not encode json.")
	}
}
