// Package middleware provides composable HTTP middleware. Each
// middleware is a function that wraps an http.Handler, making them
// trivial to test in isolation and combine in any order.
package middleware

import "net/http"

// Middleware is a http.Handler decorator.
type Middleware func(http.Handler) http.Handler

// Chain wraps h with the given middlewares. The first middleware in
// the slice becomes the outermost wrapper, so it runs first on the
// way in and last on the way out.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i](h)
	}
	return h
}
