// Package httputils provides utilities commonly used in applications serving
// HTTP requests.
//
// This root package just contains some convenience types; sub-packages add
// more utilities.
package httputils

import "net/http"

// Middleware defines the common function signature for HTTP middlewares as
// a golang type.
type Middleware = func(http.Handler) http.Handler

// Compose composes all midlewares to a single middleware. The middlewares
// are given in inner-to-outer order, so middlewares[0] will be applied first.
// This means, that
//
//	Compose(a, b)(h)
//
// is equivalent to
//
//	b(a(h))
func Compose(middlewares ...Middleware) Middleware {
	return func(h http.Handler) http.Handler {
		for _, mw := range middlewares {
			h = mw(h)
		}
		return h
	}
}
