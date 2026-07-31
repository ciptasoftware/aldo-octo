package middleware

import (
	"net/http"
)

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Chain builds a chain of middlewares around a final handler
func Chain(h http.Handler, middlewares ...Middleware) http.Handler {
	// Loop backwards through middlewares to preserve execution order
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

// ChainFunc wraps an http.HandlerFunc directly
func ChainFunc(h http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	chained := Chain(h, middlewares...)
	return func(w http.ResponseWriter, r *http.Request) {
		chained.ServeHTTP(w, r)
	}
}
