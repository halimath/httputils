package httputils

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halimath/expect"
	"github.com/halimath/expect/is"
)

func TestCompose(t *testing.T) {
	order := []string{}

	// First middleware (inner)
	mw1 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw1-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw1-after")
		})
	}

	// Second middleware (outer)
	mw2 := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			order = append(order, "mw2-before")
			next.ServeHTTP(w, r)
			order = append(order, "mw2-after")
		})
	}

	finalHandlerCalled := false
	finalHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		finalHandlerCalled = true
		order = append(order, "handler")
	})

	composed := Compose(mw1, mw2)(finalHandler)

	rr := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	composed.ServeHTTP(rr, req)

	if !finalHandlerCalled {
		t.Fatalf("final handler was not called")
	}

	expect.That(t, is.DeepEqualTo(order,
		[]string{
			"mw2-before", // outer
			"mw1-before", // inner
			"handler",
			"mw1-after", // inner
			"mw2-after", // outer
		}))
}
