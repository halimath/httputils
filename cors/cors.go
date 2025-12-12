// Package cors provides a configurable HTTP middleware to enable cross-origin resource sharing (CORS) by
// handling the respective HTTP headers and answering pre-flight requests.
package cors

import (
	"net/http"
	"strings"

	"github.com/halimath/httputils"
)

const (
	// RequestHeaderHeaders defines the Origin request header.
	RequestHeaderOrigin = "Origin"

	// RequestHeaderMethod defines the request header to issue HTTP methods.
	RequestHeaderMethod = "Access-Control-Request-Method"

	// RequestHeaderHeaders defines the request header to issue headers.
	RequestHeaderHeaders = "Access-Control-Request-Headers"

	// ResponseHeaderAllowOrigin defines the response header to signal allowed origins.
	ResponseHeaderAllowOrigin = "Access-Control-Allow-Origin"

	// ResponseHeaderAllowMethods defines the response header to signal allowed methods.
	ResponseHeaderAllowMethods = "Access-Control-Allow-Methods"

	// ResponseHeaderAllowHeaders defines the response header to signal allowed headers.
	ResponseHeaderAllowHeaders = "Access-Control-Allow-Headers"

	// ResponseHeaderAllowCredential defines the response header to signal whether credentials are allowed.
	ResponseHeaderAllowCredentials = "Access-Control-Allow-Credentials"

	// Wildcard defines the wildcard used as a value for several headers.
	Wildcard = "*"
)

// Endpoint defines how a single HTTP endpoint can be accessed in a cross-origin manner. For each configured
// Endpoint a Path must be given as it identifies the endpoint. All other struct fields can be left empty
// which marks the defaults as defined in the HTTP RFC. Set field values to customize resource sharing.
type Endpoint struct {
	// Path defines the endpoint's path and must be given.
	Path string

	// AllowMethods defines the allowed HTTP methods. If left empty, no allow methods response header is
	// sent which means that defaults apply.
	AllowMethods []string

	// AllowOrigins defines the allowed origins to access the endpoint. If left empty or set to the wildcard,
	// all origins are allowed.
	AllowOrigins []string

	// AllowHeaders lists the allowed headers for cross-origin requests. If left empty no allow headers
	// response header is sent and the defaults apply.
	AllowHeaders []string

	// AllowCredentials specifies whether credentials are allowed and the respective response header is sent.
	// If set to false (the default) the response header is not sent.
	AllowCredentials bool
}

// allowsOrigin tests whether the given origin is allowed by e. If AllowOrigins is empty of contains only the
// wildcard, every origin is allowed. Otherwise the allowed origins are compared literally.
func (e Endpoint) allowsOrigin(origin string) bool {
	if len(e.AllowOrigins) == 0 {
		return true
	}

	if len(e.AllowOrigins) == 1 && e.AllowOrigins[0] == Wildcard {
		return true
	}

	for _, o := range e.AllowOrigins {
		if o == origin {
			return true
		}
	}

	return false
}

// allEndpoint is a sentinel value used in case no endpoints are given to the middleware. This endpoint is
// then used to process requests.
var allEndpoint = Endpoint{}

// Middleware creates a HTTP middleware enabling
// Cross-Origin Resource Sharing by adding response headers and handling pre-flight requests.
// Pre-flight requests (using the HTTP method OPTIONS) are handled completely by this middleware and are not
// forwarded downstream. Other requests are forwarded to handler but HTTP response headers are set beforehand.
func Middleware(endpoints ...Endpoint) httputils.Middleware {
	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Check if the request carries an Origin header.
			if !isCrossOrigin(r) {
				// If not, simply send it downstream.
				handler.ServeHTTP(w, r)
				return
			}

			// Determine the endpoint that's applicable for the request.
			endpoint, ok := findEndpoint(r, endpoints)

			if ok {
				// If an endpoint has been configured, determine the origin.
				origin := r.Header.Get(RequestHeaderOrigin)

				if endpoint.allowsOrigin(origin) {
					// If the origin is allowed by the endpoint configuration, add the respective Allow-* headers
					// based on the configuration.
					w.Header().Add(ResponseHeaderAllowOrigin, origin)

					if len(endpoint.AllowMethods) > 0 {
						w.Header().Add(ResponseHeaderAllowMethods, strings.Join(endpoint.AllowMethods, ", "))
					}

					if len(endpoint.AllowHeaders) > 0 {
						w.Header().Add(ResponseHeaderAllowHeaders, strings.Join(endpoint.AllowHeaders, ", "))
					}

					if endpoint.AllowCredentials {
						w.Header().Add(ResponseHeaderAllowCredentials, "true")
					}
				}
			}

			if isPreflight(r) {
				// If this is a preflight request, send a response and do not send the request downstream.
				w.WriteHeader(http.StatusNoContent)
				return
			}

			// In any other way, send the request downstream.
			handler.ServeHTTP(w, r)
		})
	}
}

func findEndpoint(r *http.Request, endpoints []Endpoint) (Endpoint, bool) {
	if len(endpoints) == 0 {
		return allEndpoint, true
	}

	for _, e := range endpoints {
		if strings.HasPrefix(r.URL.Path, e.Path) {
			return e, true
		}
	}

	return allEndpoint, false
}

func isCrossOrigin(r *http.Request) bool {
	return r.Header.Get(RequestHeaderOrigin) != ""
}

func isPreflight(r *http.Request) bool {
	return r.Method == http.MethodOptions
}
