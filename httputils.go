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
