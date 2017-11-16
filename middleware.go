package jobsearch

import "net/http"

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
