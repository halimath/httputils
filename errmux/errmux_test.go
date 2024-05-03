package errmux

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halimath/expect"
	"github.com/halimath/expect/is"
	"github.com/halimath/httputils/requestbuilder"
)

func TestServeMux(t *testing.T) {
	errValue := errors.New("kaboom")

	mux := NewServeMux()

	var handledError error
	mux.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		handledError = err
		w.WriteHeader(http.StatusNotImplemented)
	}

	var requestHandled bool
	mux.HandleFunc("/err", func(_ http.ResponseWriter, _ *http.Request) error {
		requestHandled = true
		return errValue
	})

	mux.HandleFunc("/ok", func(w http.ResponseWriter, _ *http.Request) error {
		w.WriteHeader(http.StatusOK)
		return nil
	})

	recorder := httptest.NewRecorder()
	req := requestbuilder.Get("/ok").Request()
	mux.ServeHTTP(recorder, req)

	expect.That(t, is.EqualTo(recorder.Result().StatusCode, http.StatusOK))

	recorder = httptest.NewRecorder()
	req = requestbuilder.Get("/err").Request()
	mux.ServeHTTP(recorder, req)

	expect.That(t,
		is.DeepEqualTo(handledError, errValue),
		is.EqualTo(requestHandled, true),
		is.EqualTo(recorder.Result().StatusCode, http.StatusNotImplemented),
	)
}
