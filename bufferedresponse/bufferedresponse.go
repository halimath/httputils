// Package bufferedresponse provides a type that satisfies [http.ResponseWriter]
// by providing in-memory buffering. The buffered response can be used to capture
// http response data and write it to another (unbuffered) ResponseWriter. This
// enables implementations (i.e. middlewares) to discard any http response data
// and send another one (i.e. in case of an error).
package bufferedresponse

import (
	"bytes"
	"io"
	"net/http"
)

// ResponseWriter implements [http.ResponseWriter] by providing a buffer
// applications can write to. The buffer may then be written to another (i/o
// connected) ResponseWriter.
type ResponseWriter struct {
	// The status code to send
	statusCode int

	// HTTP headers and trailers to send. See [http.ResponseWriter] for an
	// explanation how to separate headers and trailers
	header http.Header

	// Body data to send
	body bytes.Buffer
}

// WriteTo writes w to target. If no status code has been set before,
// [http.StatusOK] is used. target's Header method is only invoked if w has
// headers set.
// WriteTo returns any error returned from calling target.Write.
func (w *ResponseWriter) WriteTo(target http.ResponseWriter) error {
	if len(w.header) > 0 {
		h := target.Header()
		for k, vals := range w.header {
			for _, val := range vals {
				h.Add(k, val)
			}
		}
	}

	if w.statusCode != 0 {
		target.WriteHeader(w.statusCode)
	} else {
		target.WriteHeader(http.StatusOK)
	}

	var err error
	if w.body.Len() > 0 {
		_, err = w.body.WriteTo(target)
	}

	return err
}

// Reset resets w to an empty state. Calling Reset is only possible up until
// w has been finalized. Calling Reset after Finalize causes a panic.
func (w *ResponseWriter) Reset() {
	w.header = http.Header{}
	w.body.Reset()
	w.statusCode = 0
}

// StatusCode returns the status code set for w. If no status code has been
// set so far, 0 is returned.
func (w *ResponseWriter) StatusCode() int { return w.statusCode }

// Methods implemented to satisfy http.ResponseWriter

func (w *ResponseWriter) Header() http.Header {
	if w.header == nil {
		w.header = http.Header{}
	}
	return w.header
}

func (w *ResponseWriter) Write(buf []byte) (int, error) {
	return w.body.Write(buf)
}

func (w *ResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
}

// Additional methods offered by bytes.Buffer to improve writing performance

// WriteString appends the contents of s to the body of w.
func (w *ResponseWriter) WriteString(s string) (n int, err error) {
	return w.body.WriteString(s)
}

// ReadFrom reads data from r until EOF and appends it to the buffer, growing
// the buffer as needed. The return value n is the number of bytes read. Any
// error except io.EOF encountered during the read is also returned. If the
// buffer becomes too large, ReadFrom will panic with [ErrTooLarge].
func (w *ResponseWriter) ReadFrom(r io.Reader) (n int64, err error) {
	return w.body.ReadFrom(r)
}
