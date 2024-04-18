package requesturi

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/halimath/glob"
	"github.com/halimath/httputils/internal/valuecomponents"
)

// URLRewriter is a function that updates a requests URL.
type URLRewriter func(r *http.Request)

// Middleware creates a http middleware by wrapping h in a new http.Handler
// that completes the request's URL field with protocol (scheme) and
// host, which are left empty by default. h may use r.URL to rebuild
// the full URL issued by the client when making the request.
// The list of URL rewriters may update the request's URL based on
// other request entities, such as a Forwarded-Header when running
// behind a reverse proxy.
func Middleware(h http.Handler, rewriter ...URLRewriter) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Apply defaults
		proto := "http"
		if r.TLS != nil {
			proto += "s"
		}
		r.URL.Scheme = proto
		r.URL.Host = r.Host

		for _, rw := range rewriter {
			rw(r)
		}

		h.ServeHTTP(w, r)
	})
}

const (
	// HeaderForwarded contains the key of the Forwarded request header.
	HeaderForwarded = "Forwarded"

	// SchemeHttp contains the URL scheme for HTTP.
	SchemeHttp = "http"

	// SchemeHttps contains the URL scheme for HTTPS.
	SchemeHttps = "https"
)

// Forwarded is a URLRewriter which updates URL fields based
// on Forwarded header as specified in RFC 7239
// (https://datatracker.ietf.org/doc/html/rfc7239).
// Sepcificially, the following values are always set:
//
// - r.URL.Scheme - the protocol being used (http or https)
// - r.URL.Host - the full host and port as specified by the client
func Forwarded(r *http.Request) {
	forwarded, ok := r.Header[HeaderForwarded]
	if ok && len(forwarded) > 0 {
		for _, f := range forwarded {
			applyForwarded(f, r.URL)
		}
	}

}

// applyForwarded parses the Forwarded header h as defined in
// RFC 7239 and and applies the data to u. It silently ignores any
// malformed data. See https://datatracker.ietf.org/doc/html/rfc7239
func applyForwarded(h string, u *url.URL) {
	if h == "" {
		return
	}

	vl, err := valuecomponents.ParseValueList(h)
	if err != nil {
		log.Printf("Ignoring invalid Forwarded-Header: '%s': %s", h, err)
	}

	for _, v := range vl {
		for k, v := range v.Pairs {
			switch strings.ToLower(k) {
			case "host":
				u.Host = v
			case "proto":
				proto := strings.ToLower(v)
				if proto == SchemeHttp || proto == SchemeHttps {
					u.Scheme = proto
				}
			}
		}
	}
}

const (
	// HeaderXForwardedHost contains the key for the X-Forwarded-Proto request header.
	HeaderXForwardedProto = "X-Forwarded-Proto"

	// HeaderXForwardedHost contains the key for the X-Forwarded-Host request header.
	HeaderXForwardedHost = "X-Forwarded-Host"
)

// XForwarded is a URLRewriter that rewrites the r's URL based on the X-Forwarded-*
// headers.
// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Forwarded
// for technical details.
func XForwarded(r *http.Request) {
	host, ok := r.Header[HeaderXForwardedHost]

	if ok {
		for _, h := range host {
			if len(h) > 0 {
				r.URL.Host = h
			}
		}
	}

	proto, ok := r.Header[HeaderXForwardedProto]

	if ok {
		for _, p := range proto {
			if p == SchemeHttp || p == SchemeHttps {
				r.URL.Scheme = p
			}
		}
	}
}

// RewritePath creates a URLRewriter func that rewrites URL paths based on
// mapping. The keys to mapping are compiled as patterns using
// [github.com/halimath/glob]. The values are the full paths to rewrite
// the path to.
// Patterns are compiled when RewritePath is invoked.
func RewritePath(mapping map[string]string) (URLRewriter, error) {
	patternToRewriteTarget := make(map[*glob.Pattern]string, len(mapping))
	for k, v := range mapping {
		pat, err := glob.New(k)
		if err != nil {
			return nil, fmt.Errorf("failed to compile pattern %q: %v", k, err)
		}
		patternToRewriteTarget[pat] = v
	}

	return func(r *http.Request) {
		for pat, rewriteTo := range patternToRewriteTarget {
			if pat.Match(r.URL.Path) {
				r.URL.Path = rewriteTo
				return
			}
		}
	}, nil
}
