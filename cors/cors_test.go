package cors

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halimath/expect-go"
)

var (
	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "hello world")
	})
)

func httpHeader(key, value string) expect.Matcher {
	return expect.MatcherFunc(func(ctx expect.Context, got any) {
		h, ok := got.(http.Header)
		if !ok {
			ctx.Failf("expected <%v> to be http.Header but got <%T>", got, got)
			return
		}

		v := h.Get(key)
		if value != v {
			ctx.Failf("expected HTTP header <%s> to be <%s> but got <%s>", key, value, v)
		}
	})
}

func TestMiddleware_noCorsRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	m := Middleware(h)
	m.ServeHTTP(w, r)

	expect.That(t, w.Result().StatusCode).Has(expect.Equal(http.StatusOK))
}

func TestMiddleware_corsRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware(h)
	m.ServeHTTP(w, r)

	expect.That(t, w.Result().StatusCode).Has(expect.Equal(http.StatusOK))
	expect.That(t, w.Header()).Has(httpHeader(ResponseHeaderAllowOrigin, "https://example.com"))
}

func TestMiddleware_preflightRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware(h)
	m.ServeHTTP(w, r)

	expect.That(t, w.Result().StatusCode).Has(expect.Equal(http.StatusNoContent))
	expect.That(t, w.Header()).Has(httpHeader(ResponseHeaderAllowOrigin, "https://example.com"))
}

func TestMiddleware_corsRequestWithCustomAllows(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware(h, Endpoint{
		Path:             "/",
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
		AllowHeaders:     []string{"Authorization"},
		AllowCredentials: true,
	})
	m.ServeHTTP(w, r)

	expect.That(t, w.Result().StatusCode).Has(expect.Equal(http.StatusOK))
	expect.That(t, w.Header()).Has(httpHeader(ResponseHeaderAllowOrigin, "https://example.com"))
	expect.That(t, w.Header()).Has(httpHeader(ResponseHeaderAllowMethods, "GET, POST"))
	expect.That(t, w.Header()).Has(httpHeader(ResponseHeaderAllowHeaders, "Authorization"))
	expect.That(t, w.Header()).Has(httpHeader(ResponseHeaderAllowCredentials, "true"))
}

func TestMiddleware_corsRequestWithWildcardOrigin(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware(h, Endpoint{
		Path:         "/",
		AllowOrigins: []string{Wildcard},
	})
	m.ServeHTTP(w, r)

	expect.That(t, w.Result().StatusCode).Has(expect.Equal(http.StatusOK))
	expect.That(t, w.Header()).Has(httpHeader(ResponseHeaderAllowOrigin, "https://example.com"))
}

func TestMiddleware_corsRequestWithListedOrigins(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware(h, Endpoint{
		Path:         "/",
		AllowOrigins: []string{"https://example.com", "http://example.com"},
	})
	m.ServeHTTP(w, r)

	expect.That(t, w.Result().StatusCode).Has(expect.Equal(http.StatusOK))
	expect.That(t, w.Header()).Has(httpHeader(ResponseHeaderAllowOrigin, "https://example.com"))
}

func TestMiddleware_corsRequestWithListedOriginsButNoneMatches(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware(h, Endpoint{
		Path:         "/",
		AllowOrigins: []string{"http://foobar.com"},
	})
	m.ServeHTTP(w, r)

	expect.That(t, w.Result().StatusCode).Has(expect.Equal(http.StatusOK))
	expect.That(t, w.Header()).Has(httpHeader(ResponseHeaderAllowOrigin, "")) // TODO: Rewrite with negative test
}
