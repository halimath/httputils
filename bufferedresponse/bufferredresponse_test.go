package bufferedresponse_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/halimath/expect"
	"github.com/halimath/expect/is"
	"github.com/halimath/httputils/bufferedresponse"
)

func TestBufferedResponse(t *testing.T) {
	var br bufferedresponse.ResponseWriter

	content := "hello, world"
	br.Header().Set("Content-Type", "text/plain")
	br.Header().Set("Content-Length", strconv.Itoa(len(content)))
	br.WriteHeader(http.StatusCreated)

	_, err := io.WriteString(&br, content)
	expect.That(t, is.NoError(err))

	rw := httptest.NewRecorder()
	err = br.WriteTo(rw)
	expect.That(t, is.NoError(err))

	var sb strings.Builder
	err = rw.Result().Write(&sb)
	expect.That(t,
		is.NoError(err),
		is.EqualToStringByLines(sb.String(), `HTTP/1.1 201 Created
		Content-Length: 12
		Content-Type: text/plain

		hello, world`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
	)
}
