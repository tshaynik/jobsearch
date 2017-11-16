package jobsearch

import (
	"log"
	"net/http"
	"time"
)

// Middleware can apply new behaviours to http handlers before and/or after the
// handlers are served
type Middleware interface {
	Adapt(http.Handler) http.Handler
	AdaptFunc(http.HandlerFunc) http.Handler
}

// Adapter is a function that takes an http handler and returns a modified
// http handler.
type Adapter func(http.Handler) http.Handler

// NewMiddleware chains together a collection of adapters into a middlware that
// can be applied to an http.Handler.
func NewMiddleware(as ...Adapter) Middleware {
	var result middleware
	for _, a := range as {
		result = append(result, a)
	}
	return result
}

type middleware []func(http.Handler) http.Handler

func (m middleware) Adapt(target http.Handler) http.Handler {
	for _, adapter := range m {
		target = adapter(target)
	}
	return target
}

func (m middleware) AdaptFunc(target http.HandlerFunc) http.Handler {
	handler := http.Handler(target)
	return m.Adapt(handler)
}

// LogRequest takes a logger and returns an Adapter that logs http request details.
func LogRequest(l *log.Logger) Adapter {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			next.ServeHTTP(w, r)
			duration := time.Since(start)
			l.Printf(" %s | %s %s\t | user: %s | duration: %s\n",
				r.Host, r.Method, r.URL.Path, duration,
				r.Context().Value(contextKey("login")).(string))
		})
	}
}
