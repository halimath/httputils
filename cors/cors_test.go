package cors

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halimath/expect"
	"github.com/halimath/expect/is"
)

var (
	h = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "hello world")
	})
)

func hasHTTPHeader(got http.Header, key, value string) expect.ExpectFunc {
	return expect.ExpectFunc(func(t expect.TB) {
		v := got.Get(key)
		if value != v {
			t.Errorf("expected HTTP header <%s> to be <%s> but got <%s>", key, value, v)
		}
	})
}

func TestMiddleware_noCorsRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	m := Middleware()(h)
	m.ServeHTTP(w, r)

	expect.That(t, is.EqualTo(w.Result().StatusCode, http.StatusOK))
}

func TestMiddleware_corsRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware()(h)
	m.ServeHTTP(w, r)

	expect.That(t,
		is.EqualTo(w.Result().StatusCode, http.StatusOK),
		hasHTTPHeader(w.Header(), ResponseHeaderAllowOrigin, "https://example.com"),
	)
}

func TestMiddleware_preflightRequest(t *testing.T) {
	r := httptest.NewRequest(http.MethodOptions, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware()(h)
	m.ServeHTTP(w, r)

	expect.That(t,
		is.EqualTo(w.Result().StatusCode, http.StatusNoContent),
		hasHTTPHeader(w.Header(), ResponseHeaderAllowOrigin, "https://example.com"),
	)
}

func TestMiddleware_corsRequestWithCustomAllows(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware(Endpoint{
		Path:             "/",
		AllowMethods:     []string{http.MethodGet, http.MethodPost},
		AllowHeaders:     []string{"Authorization"},
		AllowCredentials: true,
	})(h)
	m.ServeHTTP(w, r)

	expect.That(t,
		is.EqualTo(w.Result().StatusCode, http.StatusOK),
		hasHTTPHeader(w.Header(), ResponseHeaderAllowOrigin, "https://example.com"),
		hasHTTPHeader(w.Header(), ResponseHeaderAllowMethods, "GET, POST"),
		hasHTTPHeader(w.Header(), ResponseHeaderAllowHeaders, "Authorization"),
		hasHTTPHeader(w.Header(), ResponseHeaderAllowCredentials, "true"),
	)
}

func TestMiddleware_corsRequestWithWildcardOrigin(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware(Endpoint{
		Path:         "/",
		AllowOrigins: []string{Wildcard},
	})(h)
	m.ServeHTTP(w, r)

	expect.That(t,
		is.EqualTo(w.Result().StatusCode, http.StatusOK),
		hasHTTPHeader(w.Header(), ResponseHeaderAllowOrigin, "https://example.com"),
	)
}

func TestMiddleware_corsRequestWithListedOrigins(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware(Endpoint{
		Path:         "/",
		AllowOrigins: []string{"https://example.com", "http://example.com"},
	})(h)
	m.ServeHTTP(w, r)

	expect.That(t,
		is.EqualTo(w.Result().StatusCode, http.StatusOK),
		hasHTTPHeader(w.Header(), ResponseHeaderAllowOrigin, "https://example.com"),
	)
}

func TestMiddleware_corsRequestWithListedOriginsButNoneMatches(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	r.Header.Add(RequestHeaderOrigin, "https://example.com")
	w := httptest.NewRecorder()

	m := Middleware(Endpoint{
		Path:         "/",
		AllowOrigins: []string{"http://foobar.com"},
	})(h)
	m.ServeHTTP(w, r)

	expect.That(t,
		is.EqualTo(w.Result().StatusCode, http.StatusOK),
		hasHTTPHeader(w.Header(), ResponseHeaderAllowOrigin, ""),
	)
}
