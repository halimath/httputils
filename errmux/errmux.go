// Package errmux provides types that extend [http.ServeMux] with error handling
// capabilities.
//
// ServeMux as defined by this package works exactly the same as [http.ServeMux]
// with the exception, that [Handler] and [HandlerFunc] respectively return an
// error value. Any non-nil error causes the response to be discarded and the
// error gets handled.
package errmux

import (
	"net/http"

	"github.com/halimath/httputils/bufferedresponse"
	"github.com/halimath/httputils/response"
)

// Handler defines an extension of [http.Handler] that returns and error value
// which causes normal response handling to be aborted and the error being
// handled.
type Handler interface {
	// ServeHTTP serves an [http.Request] producing response to an
	// [http.ResponseWriter]. The writer is buffered and any data written may
	// be discarded if ServeHTTP returns a non-nil error.
	ServeHTTP(http.ResponseWriter, *http.Request) error
}

// HandlerFunc is a convenience function type to implement [Handler] with just
// a function.
type HandlerFunc func(http.ResponseWriter, *http.Request) error

func (h HandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) error {
	return h(w, r)
}

// ErrorHandler defines a function type for handling errors that occured during
// request handling. The function is passed the original [http.ResponseWriter]
// and error handling happens unbuffered.
type ErrorHandler func(http.ResponseWriter, *http.Request, error)

// defaultErrorHandler is the default error handler which uses [response.Error]
// to send an error.
func defaultErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	response.Error(w, r, err)
}

// ServeMux works like a [http.ServeMux] but with support for error-aware request
// handling.
type ServeMux struct {
	mux          *http.ServeMux
	ErrorHandler ErrorHandler
}

func NewServeMux() *ServeMux {
	return &ServeMux{
		mux:          http.NewServeMux(),
		ErrorHandler: defaultErrorHandler,
	}
}

// decorate is used to decorate h with response writer buffering and error
// dispatching. The resulting [http.Handler] is registered with a [http.ServeMux].
func (mux *ServeMux) decorate(h Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var buf bufferedresponse.ResponseWriter
		err := h.ServeHTTP(&buf, r)

		if err == nil {
			buf.WriteTo(w)
			return
		}

		h := mux.ErrorHandler
		if h == nil {
			h = defaultErrorHandler
		}

		h(w, r, err)
	})
}

// ServeHTTP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (mux *ServeMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	mux.mux.ServeHTTP(w, r)
}

// Handle registers the handler for the given pattern.
// If the given pattern conflicts, with one that is already registered, Handle
// panics.
func (mux *ServeMux) Handle(pattern string, handler Handler) {
	mux.mux.Handle(pattern, mux.decorate(handler))
}

// HandleFunc registers the handler function for the given pattern.
// If the given pattern conflicts, with one that is already registered, HandleFunc
// panics.
func (mux *ServeMux) HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request) error) {
	mux.Handle(pattern, HandlerFunc(handler))
}
