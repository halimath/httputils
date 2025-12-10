package securityheader_test

import (
	"net/http"

	"github.com/halimath/httputils/securityheader"
)

func Example() {
	var h http.Handler

	// ...

	h = securityheader.Middleware(
		securityheader.ContentSecurityPolicy(
			securityheader.CSPPolicyDirective(securityheader.CSPDefaultSrc, securityheader.CSPSelf)),
		securityheader.StrictTransportSecurity(),
	)(h)
}
