package requesturi

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
	"github.com/halimath/httputils/requestbuilder"
)

func TestURL(t *testing.T) {
	table := map[*http.Request]string{
		requestbuilder.Get("http://http.host/foo/bar").
			Request(): "http://http.host/foo/bar",

		requestbuilder.Get("https://https.host/foo/bar").
			Request(): "https://https.host/foo/bar",
	}

	for r, exp := range table {
		var w httptest.ResponseRecorder

		f := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if diff := deep.Equal(exp, r.URL.String()); diff != nil {
				t.Error(diff)
			}
		}))

		f.ServeHTTP(&w, r)
	}
}

func TestForwarded(t *testing.T) {
	table := map[*http.Request]string{
		requestbuilder.Get("http://no.header/foo/bar").
			Request(): "http://no.header/foo/bar",

		requestbuilder.Get("http://empty.header/foo/bar").
			AddHeader("Forwarded", "").
			Request(): "http://empty.header/foo/bar",

		requestbuilder.Get("http://forwarded.header/foo/bar").
			AddHeader("Forwarded", "proto=https;host=localhost;for=1.2.3.4,for=9.8.7.6").
			Request(): "https://localhost/foo/bar",

		requestbuilder.Get("https://forwarded.header/foo/bar").
			AddHeader("Forwarded", "proto=http;host=localhost;for=1.2.3.4,for=9.8.7.6").
			Request(): "http://localhost/foo/bar",

		requestbuilder.Get("http://multiple.forwarded.header/foo/bar").
			AddHeader("Forwarded", "proto=https").
			AddHeader("Forwarded", "host=localhost").
			AddHeader("Forwarded", "for=1.2.3.4, for=9.8.7.6").
			Request(): "https://localhost/foo/bar",

		requestbuilder.Get("http://invalid.header/foo/bar").
			AddHeader("Forwarded", "proto=https,host=localhost,foo=====99=").
			Request(): "http://invalid.header/foo/bar",
	}

	for r, exp := range table {
		var w httptest.ResponseRecorder

		f := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if diff := deep.Equal(exp, r.URL.String()); diff != nil {
				t.Error(diff)
			}
		}), Forwarded)

		f.ServeHTTP(&w, r)
	}
}

func TestXForwarded(t *testing.T) {
	table := map[*http.Request]string{
		requestbuilder.Get("http://no.header/foo/bar").
			Request(): "http://no.header/foo/bar",

		requestbuilder.Get("http://empty.header/foo/bar").
			AddHeader(HeaderXForwardedHost, "").
			Request(): "http://empty.header/foo/bar",

		requestbuilder.Get("http://forwarded-host.header/foo/bar").
			AddHeader(HeaderXForwardedHost, "localhost").
			Request(): "http://localhost/foo/bar",

		requestbuilder.Get("https://forwarded-proto.header/foo/bar").
			AddHeader(HeaderXForwardedProto, "http").
			Request(): "http://forwarded-proto.header/foo/bar",
	}

	for r, exp := range table {
		var w httptest.ResponseRecorder

		f := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if diff := deep.Equal(exp, r.URL.String()); diff != nil {
				t.Error(diff)
			}
		}), XForwarded)

		f.ServeHTTP(&w, r)
	}
}
