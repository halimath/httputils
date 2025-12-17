package session

import (
	"crypto/tls"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halimath/expect"
	"github.com/halimath/expect/is"
)

func TestMiddleware(t *testing.T) {

	t.Run("withCookieOption", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		mw := NewMiddleware(WithCookieOptions(CookieOpts{Name: "mycookie"}))(h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		mw.ServeHTTP(rw, req)

		cookies := rw.Result().Cookies()
		expect.That(t,
			is.EqualTo("mycookie", cookies[0].Name),
			is.SliceOfLen(rw.Result().Cookies(), 1),
		)
	})

	t.Run("withTLS", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		mw := NewMiddleware(WithCookieOptions(CookieOpts{Name: "mycookie"}))(h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.TLS = new(tls.ConnectionState)

		mw.ServeHTTP(rw, req)

		cookies := rw.Result().Cookies()
		expect.That(t,
			is.EqualTo("mycookie", cookies[0].Name),
			is.SliceOfLen(rw.Result().Cookies(), 1),
		)
	})

	t.Run("withDefaultStore", func(t *testing.T) {
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})

		mw := NewMiddleware()(h)

		// Assert session cookie is created
		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		mw.ServeHTTP(rw, req)

		expect.That(t, is.SliceOfLen(rw.Result().Cookies(), 1))
	})

	t.Run("readExistingSession", func(t *testing.T) {
		store := NewInMemoryStore()
		ses, err := store.Create()
		expect.That(t, is.NoError(err))
		ses.Set("foo", "bar")

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ses := FromContext(r.Context())
			got := ses.Get("foo")
			expect.That(t,
				is.EqualTo(got, "bar"),
			)
		})

		mw := NewMiddleware(WithStore(store), WithCookieOptions(CookieOpts{
			Name: "sid",
		}))(h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "sid", Value: ses.ID()})

		mw.ServeHTTP(rw, req)
	})

	t.Run("createNewSessionIfMissing", func(t *testing.T) {
		store := NewInMemoryStore()

		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ses := FromContext(r.Context())
			expect.That(t, is.StringOfLen(ses.ID(), 43))
		})

		mw := NewMiddleware(WithStore(store), WithCookieOptions(CookieOpts{
			Name: "sid",
		}))(h)

		rw := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/", nil)

		mw.ServeHTTP(rw, req)

		// Cookie should be set
		expect.That(t, is.SliceOfLen(rw.Result().Cookies(), 1))
	})
}
