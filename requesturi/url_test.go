package requesturi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halimath/expect"
	"github.com/halimath/expect/is"
	"github.com/halimath/httputils/requestbuilder"
)

func TestMiddleware(t *testing.T) {
	table := map[*http.Request]string{
		requestbuilder.Get("http://http.host/foo/bar").
			Request(): "http://http.host/foo/bar",

		requestbuilder.Get("https://https.host/foo/bar").
			Request(): "https://https.host/foo/bar",
	}

	testMiddlewareWithRewriters(t, table)
}

func TestForwarded(t *testing.T) {
	table := map[*http.Request]string{
		requestbuilder.Get("http://no.header/foo/bar").
			Request(): "http://no.header/foo/bar",

		requestbuilder.Get("http://empty.header/foo/bar").
			AddHeader("Forwarded", "").
			Request(): "http://empty.header/foo/bar",

		requestbuilder.Get("http://forwarded.header/foo/bar").
			AddHeader("Forwarded", "proto=https;host=localhost;for=1.2.3.4,for=9.8.7.6").
			Request(): "https://localhost/foo/bar",

		requestbuilder.Get("https://forwarded.header/foo/bar").
			AddHeader("Forwarded", "proto=http;host=localhost;for=1.2.3.4,for=9.8.7.6").
			Request(): "http://localhost/foo/bar",

		requestbuilder.Get("http://multiple.forwarded.header/foo/bar").
			AddHeader("Forwarded", "proto=https").
			AddHeader("Forwarded", "host=localhost").
			AddHeader("Forwarded", "for=1.2.3.4, for=9.8.7.6").
			Request(): "https://localhost/foo/bar",

		requestbuilder.Get("http://invalid.header/foo/bar").
			AddHeader("Forwarded", "proto=https,host=localhost,foo=====99=").
			Request(): "http://invalid.header/foo/bar",
	}

	testMiddlewareWithRewriters(t, table, Forwarded)
}

func TestXForwarded(t *testing.T) {
	table := map[*http.Request]string{
		requestbuilder.Get("http://no.header/foo/bar").
			Request(): "http://no.header/foo/bar",

		requestbuilder.Get("http://empty.header/foo/bar").
			AddHeader(HeaderXForwardedHost, "").
			Request(): "http://empty.header/foo/bar",

		requestbuilder.Get("http://forwarded-host.header/foo/bar").
			AddHeader(HeaderXForwardedHost, "localhost").
			Request(): "http://localhost/foo/bar",

		requestbuilder.Get("https://forwarded-proto.header/foo/bar").
			AddHeader(HeaderXForwardedProto, "http").
			Request(): "http://forwarded-proto.header/foo/bar",
	}

	testMiddlewareWithRewriters(t, table, XForwarded)
}

func TestRewritePath(t *testing.T) {
	table := map[*http.Request]string{
		// Match first pattern
		requestbuilder.Get("http://example.com/foo/bar").
			Request(): "http://example.com/some/path",

		// Match second pattern
		requestbuilder.Get("http://example.com/spam/bar").
			Request(): "http://example.com/another/path",

		// No match in pattern
		requestbuilder.Get("http://example.com/not/matched").
			Request(): "http://example.com/not/matched",

		// Too deep to match pattern
		requestbuilder.Get("http://example.com/spam/intermediate/final").
			Request(): "http://example.com/spam/intermediate/final",
	}

	rewriter, err := RewritePath(map[string]string{
		"/foo/**/*": "/some/path",
		"/spam/*":   "/another/path",
	})

	expect.That(t, expect.FailNow(is.NoError(err)))

	testMiddlewareWithRewriters(t, table, rewriter)
}

func testMiddlewareWithRewriters(t *testing.T, table map[*http.Request]string, rewriteFuncs ...URLRewriter) {
	t.Helper()

	for r, want := range table {
		var w httptest.ResponseRecorder

		f := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expect.That(t, is.EqualTo(r.URL.String(), want))
		}), rewriteFuncs...)

		f.ServeHTTP(&w, r)
	}
}
