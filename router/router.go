package router

import (
	"net/http"
)

// HandlerFunc is our custom handler type that could eventually support returning errors.
type HandlerFunc func(w http.ResponseWriter, r *http.Request)

// Router is a lightweight wrapper around http.ServeMux
type Router struct {
	mux *http.ServeMux
}

// New creates a new Router
func New() *Router {
	return &Router{
		mux: http.NewServeMux(),
	}
}

// ServeHTTP makes the Router implement the http.Handler interface
func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mux.ServeHTTP(w, r)
}

// Get registers a GET route
func (rt *Router) Get(path string, handler HandlerFunc) {
	rt.mux.HandleFunc("GET "+path, handler)
}

// Post registers a POST route
func (rt *Router) Post(path string, handler HandlerFunc) {
	rt.mux.HandleFunc("POST "+path, handler)
}

// Put registers a PUT route
func (rt *Router) Put(path string, handler HandlerFunc) {
	rt.mux.HandleFunc("PUT "+path, handler)
}

// Delete registers a DELETE route
func (rt *Router) Delete(path string, handler HandlerFunc) {
	rt.mux.HandleFunc("DELETE "+path, handler)
}

// Handle allows registering custom methods or wildcard routes
func (rt *Router) Handle(pattern string, handler http.Handler) {
	rt.mux.Handle(pattern, handler)
}
