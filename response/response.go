// Package response provides methods to send HTTP responses in different formats.
package response

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/textproto"
	"runtime"
	"strconv"
	"strings"

	"github.com/halimath/httputils/bufferedresponse"
)

// Option defines a function type used to customize a response. Any non-nil error
// returned from an Option causes the response to be aborted.
type Option func(w http.ResponseWriter, r *http.Request) error

// StatusCode is an Option that sets a response' HTTP status code. Only the first
// StatusCode Option applied to a response takes effect. All following calls are
// no-ops.
func StatusCode(sc int) Option {
	return func(w http.ResponseWriter, r *http.Request) error {
		if gsc, ok := w.(interface{ StatusCode() int }); ok {
			if gsc.StatusCode() != 0 {
				return nil
			}
		}

		w.WriteHeader(sc)
		return nil
	}
}

// AddHeader returns an Option that adds a header h with value v to a response.
// It uses the [http.Header.Add] method.
func AddHeader(h, v string) Option {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Add(h, v)
		return nil
	}
}

// SetHeader returns an Option that sets a header h with value v to a response.
// It uses the [http.Header.Set] method. overwrite defines whether SetHeader
// should overwrite a header value for h set previously or if should skip
// setting v.
func SetHeader(h, v string, overwwrite bool) Option {
	h = textproto.CanonicalMIMEHeaderKey(h)
	return func(w http.ResponseWriter, r *http.Request) error {
		_, ok := w.Header()[h]
		if ok && !overwwrite {
			return nil
		}

		w.Header().Set(h, v)
		return nil
	}
}

// WriteBody returns an Option that writes buf as the reponse's body.
func WriteBody(buf []byte) Option {
	return func(w http.ResponseWriter, r *http.Request) error {
		_, err := w.Write(buf)
		return err
	}
}

// Send is a generic response sending operation. It applies all opts and writes
// the response to w. It returns any technical (i.e. i/o) error that occured.
// Usually, these errors should be logged but need no further treatment.
func Send(w http.ResponseWriter, r *http.Request, opts ...Option) error {
	var buf bufferedresponse.ResponseWriter
	for _, opt := range opts {
		if err := opt(&buf, r); err != nil {
			return err
		}
	}

	return buf.WriteTo(w)
}

// DevMode can be set to true to enable error responses for development/debugging. Otherwise, errors are
// discarded.
var DevMode = false

// Error sends an error response. By default, a status code 500
// [http.StatusInternalServerError] is used but opts may replace this with a
// different status code.
//
// This operation pays respect to DevMode. If DevMode is false (the default),
// Error simply sends an empty response with the respective status code. If
// DevMode is set to true, this method sends the errors description as a plain
// text response simplifying development.
func Error(w http.ResponseWriter, r *http.Request, err error, opts ...Option) error {
	if DevMode {
		return PlainText(w, r, buildErrorResponse(err), append(opts, StatusCode(http.StatusInternalServerError))...)
	}

	return Send(w, r, append(opts, StatusCode(http.StatusInternalServerError))...)
}

func buildErrorResponse(err error) string {
	pc := make([]uintptr, 256)
	runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc)

	var sb strings.Builder

	fmt.Fprintf(&sb, "%s (%T)\n", err.Error(), err)
	for {
		f, ok := frames.Next()
		if !ok {
			break
		}
		fmt.Fprintf(&sb, "at %s (%s:%d)\n", f.Function, f.File, f.Line)
	}

	return sb.String()
}

// Forbidden sends a plain text response with status code 403 ([http.StatusForbidden]).
// opts may customize headers.
func Forbidden(w http.ResponseWriter, r *http.Request, opts ...Option) error {
	return PlainText(w, r, http.StatusText(http.StatusForbidden), append(opts, StatusCode(http.StatusForbidden))...)
}

// NotFound sends a plain text response with status code 404 ([http.StatusNotFound]).
// opts may customize headers.
func NotFound(w http.ResponseWriter, r *http.Request, opts ...Option) error {
	return PlainText(w, r, http.StatusText(http.StatusNotFound), append(opts, StatusCode(http.StatusNotFound))...)
}

// PlainText sends a response with a plain text body. It sends body as the
// response' body. opts may further customize the response (headers, status code).
// PlainText sets content-type to text/plain (overwritable) and content-length
// to the body's length (not overwritable).
func PlainText(w http.ResponseWriter, r *http.Request, body string, opts ...Option) error {
	return Send(w, r, append(opts,
		SetHeader("Content-Type", "text/plain", false),
		SetHeader("Content-Length", strconv.Itoa(len(body)), true),
		WriteBody([]byte(body)),
	)...)
}

// NotModified sends an empty response with status 304 ([http.StatusNotModified]).
// The status code may be changed.
func NotModified(w http.ResponseWriter, r *http.Request, opts ...Option) error {
	return NoContent(w, r, append(opts, StatusCode(http.StatusNotModified))...)
}

// NoContent sends an empty response with status 204 ([http.StatusNoContent]) by
// default. The content-length header is set to 0 and cannot be overwritten.
// opts may changed the status code.
func NoContent(w http.ResponseWriter, r *http.Request, opts ...Option) error {
	return Send(w, r, append(opts, SetHeader("Content-Length", "0", true), StatusCode(http.StatusNoContent))...)
}

// JSON sends a response with content-type application/json. It uses [json.Marshal]
// to marshal payload to JSON bytes and sends them as the response' body. JSON
// sets both header content-type and content-length. opts may further customize
// the response as well as overwrite the content-type (to some other content
// type based on JSON).
//
// If DevMode is set to true, JSON output is pretty printed
func JSON(w http.ResponseWriter, r *http.Request, payload any, opts ...Option) error {
	var data []byte
	var err error

	if DevMode {
		data, err = json.MarshalIndent(payload, "", "  ")
	} else {
		data, err = json.Marshal(payload)
	}
	if err != nil {
		return Error(w, r, err)
	}

	return Send(w, r, append(opts,
		SetHeader("Content-Type", "application/json", false),
		SetHeader("Content-Length", strconv.Itoa(len(data)), true),
		WriteBody(data),
	)...)
}

// ProblemDetails defines a problem details object as defined by [RFC9457].
// ProblemDetails defines an Errors field which may be used to deliver additional
// error information as an extension.
//
// [RFC9457]: https://www.rfc-editor.org/rfc/rfc9457
type ProblemDetails struct {
	// Type discriminator - must be given
	Type string `json:"type"`

	// Human readable title - must be given
	Title string `json:"title"`

	// Status code - may be set. If set, also defines the HTTP status code
	Status int `json:"status,omitempty"`

	// Additional human readable details - optional
	Detail string `json:"detail,omitempty"`

	// Identifier pointing to the instance that caused this problem - optional
	Instance string `json:"instance,omitempty"`

	// Additional user defined error information - optional and used as an extension
	Errors []any `json:"errors,omitempty"`
}

// Problem sends problemDetails as a JSON response as defined by [RFC9457].
//
// [RFC9457]: https://www.rfc-editor.org/rfc/rfc9457
func Problem(w http.ResponseWriter, r *http.Request, problemDetails ProblemDetails, opts ...Option) error {
	status := http.StatusInternalServerError

	if problemDetails.Status != 0 {
		status = problemDetails.Status
	}

	return JSON(w, r, problemDetails, SetHeader("Content-Type", "application/problem+json", true), StatusCode(status))
}
