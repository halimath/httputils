package securityheader

import "net/http"

// A middleware Option to set the X-Content-Type-Options header to noniff -
// the only supported directive for this header.
func XContentTypeOptions(h http.Header) {
	h.Set("X-Content-Type-Options", "nosniff")
}

// --

type xFrameOptionsDirective string

const (
	// The page cannot be displayed in a frame, regardless of the site attempting to do so. Not only will the browser attempt to load the page in a frame fail when loaded from other sites, attempts to do so will fail when loaded from the same site.
	XFrameOptionsDirectiveDeny xFrameOptionsDirective = "DENY"

	// The page can only be displayed if all ancestor frames have the same origin as the page itself. You can still use the page in a frame as long as the site including it in a frame is the same as the one serving the page.
	XFrameOptionsDirectiveSameOrigin xFrameOptionsDirective = "SAMEORIGIN"
)

// A middleware Option that sets the X-Frame-Options header to directive.
func XFrameOptions(directive xFrameOptionsDirective) Option {
	return func(h http.Header) {
		h.Set("X-Frame-Options", string(directive))
	}
}
