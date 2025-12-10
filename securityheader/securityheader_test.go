package securityheader

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/halimath/expect"
	"github.com/halimath/expect/is"
	"github.com/halimath/httputils"
)

func executeMW(mw httputils.Middleware) http.Header {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	wrapped := mw(h)

	req := httptest.NewRequest("GET", "/", nil)
	res := new(httptest.ResponseRecorder)

	wrapped.ServeHTTP(res, req)

	return res.Header()
}

func TestMiddleware(t *testing.T) {
	t.Run("Content-Security-Policy", func(t *testing.T) {
		t.Run("default options", func(t *testing.T) {
			hdr := executeMW(Middleware(ContentSecurityPolicy()))

			expect.That(t,
				is.EqualTo(hdr.Get("Content-Security-Policy"), "default-src 'self'"),
			)
		})
		t.Run("single custom option", func(t *testing.T) {
			hdr := executeMW(Middleware(ContentSecurityPolicy(CSPPolicyDirective(CSPScriptSrc, CSPNone))))

			expect.That(t,
				is.EqualTo(hdr.Get("Content-Security-Policy"), "script-src 'none'"),
			)
		})
		t.Run("multiple custom options", func(t *testing.T) {
			hdr := executeMW(Middleware(ContentSecurityPolicy(
				CSPPolicyDirective(CSPScriptSrc, CSPNone),
				CSPPolicyDirective(CSPImgSrc, CSPSelf),
				CSPPolicyDirective(CSPObjectSrc, "https:"),
			)))

			expect.That(t,
				is.EqualTo(hdr.Get("Content-Security-Policy"), "script-src 'none'; img-src 'self'; object-src https:"),
			)
		})
	})

	t.Run("Strict-Transport-Security", func(t *testing.T) {
		t.Run("default options", func(t *testing.T) {
			hdr := executeMW(Middleware(StrictTransportSecurity()))

			expect.That(t,
				is.EqualTo(hdr.Get("Strict-Transport-Security"), "max-age=31536000"),
			)
		})
		t.Run("single custom option", func(t *testing.T) {
			hdr := executeMW(Middleware(StrictTransportSecurity(HSTSIncludeSubDomains)))

			expect.That(t,
				is.EqualTo(hdr.Get("Strict-Transport-Security"), "includeSubDomains"),
			)
		})
		t.Run("multiple custom options", func(t *testing.T) {
			hdr := executeMW(Middleware(StrictTransportSecurity(
				HSTSMaxAge(time.Hour),
				HSTSIncludeSubDomains,
				HSTSPreload,
			)))

			expect.That(t,
				is.EqualTo(hdr.Get("Strict-Transport-Security"), "max-age=3600; includeSubDomains; preload"),
			)
		})
	})

	t.Run("X-Content-Type-Options", func(t *testing.T) {
		hdr := executeMW(Middleware(XContentTypeOptions))

		expect.That(t,
			is.EqualTo(hdr.Get("X-Content-Type-Options"), "nosniff"),
		)
	})

	t.Run("X-Frame-Options", func(t *testing.T) {
		hdr := executeMW(Middleware(XFrameOptions(XFrameOptionsDirectiveDeny)))

		expect.That(t,
			is.EqualTo(hdr.Get("X-Frame-Options"), "DENY"),
		)
	})
}
