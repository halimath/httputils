// Package securityheader provides a http middleware to inject security related
// response headers.
// The package currently supports the following header:
//
// - Content-Security-Policy
package securityheader

import (
	"net/http"
	"strings"

	"github.com/halimath/httputils"
)

// An option to customize security header.
type Option func(http.Header)

// Middleware defines a HTTP middleware that injects the security headers given
// via opts.
func Middleware(opts ...Option) httputils.Middleware {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			responseHeader := w.Header()

			for _, opt := range opts {
				opt(responseHeader)
			}

			h.ServeHTTP(w, r)
		})
	}
}

type directive interface {
	headerValue() string
}

func joinDirectives[T directive](directives []T) string {
	var b strings.Builder
	for i, directive := range directives {
		if i > 0 {
			b.WriteString("; ")
		}
		b.WriteString(directive.headerValue())
	}

	return b.String()
}
