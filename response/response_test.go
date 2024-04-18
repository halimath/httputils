package response

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"

	"github.com/halimath/expect"
	"github.com/halimath/expect/is"
)

func TestSend(t *testing.T) {
	t.Run("no_opts", func(t *testing.T) {
		got, err := apply(func(w http.ResponseWriter, r *http.Request) error { return Send(w, r) })
		expect.That(t,
			is.NoError(err),
			is.EqualToStringByLines(got, `HTTP/1.1 200 OK
			Connection: close
			
			`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
		)
	})

	t.Run("StatusCode", func(t *testing.T) {
		got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
			return Send(w, r, StatusCode(http.StatusAccepted))
		})
		expect.That(t,
			is.NoError(err),
			is.EqualToStringByLines(got, `HTTP/1.1 202 Accepted
			Connection: close
			
			`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
		)
	})

	t.Run("AddHeader", func(t *testing.T) {
		got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
			return Send(w, r, AddHeader("Foo", "bar"))
		})
		expect.That(t,
			is.NoError(err),
			is.EqualToStringByLines(got, `HTTP/1.1 200 OK
			Connection: close
			Foo: bar
			
			`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
		)
	})

	t.Run("SetHeader_noOverwrite", func(t *testing.T) {
		got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
			return Send(w, r, AddHeader("Foo", "bar"), SetHeader("Foo", "spam", false))
		})
		expect.That(t,
			is.NoError(err),
			is.EqualToStringByLines(got, `HTTP/1.1 200 OK
			Connection: close
			Foo: bar
			
			`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
		)
	})

	t.Run("SetHeader_overwrite", func(t *testing.T) {
		got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
			return Send(w, r, AddHeader("Foo", "bar"), SetHeader("Foo", "spam", true))
		})
		expect.That(t,
			is.NoError(err),
			is.EqualToStringByLines(got, `HTTP/1.1 200 OK
			Connection: close
			Foo: spam
			
			`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
		)
	})

	t.Run("WriteBody", func(t *testing.T) {
		got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
			return Send(w, r, WriteBody([]byte("hello")))
		})
		expect.That(t,
			is.NoError(err),
			is.EqualToStringByLines(got, `HTTP/1.1 200 OK
			Connection: close
			
			hello`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
		)
	})
}

func TestError(t *testing.T) {
	t.Run("no_devmode", func(t *testing.T) {
		got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
			return Error(w, r, fmt.Errorf("kaboom"))
		})
		expect.That(t,
			is.NoError(err),
			is.EqualToStringByLines(got, `HTTP/1.1 500 Internal Server Error
			Connection: close
			
			`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
		)
	})

	t.Run("devmode", func(t *testing.T) {
		DevMode = true
		defer func() {
			DevMode = false
		}()

		var frames *runtime.Frames

		got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
			pc := make([]uintptr, 256)
			runtime.Callers(1, pc)
			frames = runtime.CallersFrames(pc)

			return Error(w, r, fmt.Errorf("kaboom"))
		})

		body := "kaboom (*errors.errorString)\n"
		first := true
		for {
			f, ok := frames.Next()
			if !ok {
				break
			}
			if first {
				f.Line += 3
				first = false
			}
			body += fmt.Sprintf("at %s (%s:%d)\n", f.Function, f.File, f.Line)
		}

		want := fmt.Sprintf(`HTTP/1.1 500 Internal Server Error
		Content-Length: %d
		Content-Type: text/plain

		%s`, len(body), body)

		expect.That(t,
			is.NoError(err),
			is.EqualToStringByLines(got, want, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
		)
	})
}

func TestNotModified(t *testing.T) {
	got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
		return NotModified(w, r)
	})

	expect.That(t,
		is.NoError(err),
		is.EqualToStringByLines(got, `HTTP/1.1 304 Not Modified
		
		`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
	)
}

func TestNoContent(t *testing.T) {
	got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
		return NoContent(w, r)
	})

	expect.That(t,
		is.NoError(err),
		is.EqualToStringByLines(got, `HTTP/1.1 204 No Content
		
		`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
	)
}

func TestJSON(t *testing.T) {
	payload := struct {
		Foo string `json:"foo"`
		Bar int    `json:"bar"`
	}{Foo: "foo", Bar: 17}

	t.Run("noDevMode", func(t *testing.T) {
		got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
			return JSON(w, r, payload)
		})

		expect.That(t,
			is.NoError(err),
			is.EqualToStringByLines(got, `HTTP/1.1 200 OK
			Content-Length: 22
			Content-Type: application/json

			{"foo":"foo","bar":17}`, is.DedentLines, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
		)
	})

	t.Run("DevMode", func(t *testing.T) {
		DevMode = true
		defer func() { DevMode = false }()

		got, err := apply(func(w http.ResponseWriter, r *http.Request) error {
			return JSON(w, r, payload)
		})

		expect.That(t,
			is.NoError(err),
			is.EqualToStringByLines(got, `HTTP/1.1 200 OK
Content-Length: 31
Content-Type: application/json

{
  "foo": "foo",
  "bar": 17
}`, func(s string) string { return strings.ReplaceAll(s, "\r", "") }),
		)

	})
}

func apply(f func(w http.ResponseWriter, r *http.Request) error) (string, error) {
	w := httptest.NewRecorder()
	r := httptest.NewRequest("get", "https://example.com/some/path", nil)

	if err := f(w, r); err != nil {
		return "", err
	}

	var sb strings.Builder
	if err := w.Result().Write(&sb); err != nil {
		return "", err
	}

	return sb.String(), nil
}
