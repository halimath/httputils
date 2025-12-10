package securityheader

import (
	"fmt"
	"net/http"
	"time"
)

type hstsDirective string

func (d hstsDirective) headerValue() string { return string(d) }

const (
	// If this directive is specified, the HSTS policy applies to all subdomains of the host's domain as well.
	HSTSIncludeSubDomains hstsDirective = "includeSubDomains"

	// See Preloading Strict Transport Security for details. When using preload,
	// the max-age directive must be at least 31536000 (1 year), and the includeSubDomains directive must be present.
	HSTSPreload hstsDirective = "preload"
)

// The time, in seconds, that the browser should remember that a host is only to be accessed using HTTPS.
func HSTSMaxAge(dur time.Duration) hstsDirective {
	return hstsDirective(fmt.Sprintf("max-age=%.0f", dur.Seconds()))
}

// Returns a middleware Option that sets the Strict-Transport-Security header
// based on directives. If no directives are given,
//
//	max-age=31536000
//
// is used.
func StrictTransportSecurity(directives ...hstsDirective) Option {
	if len(directives) == 0 {
		directives = append(directives, HSTSMaxAge(31536000*time.Second))
	}

	value := joinDirectives(directives)

	return func(h http.Header) {
		h.Set("Strict-Transport-Security", value)
	}
}
