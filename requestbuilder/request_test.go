package requestbuilder

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
)

func TestRequestBuilder(t *testing.T) {
	var body bytes.Reader

	reqWithHeader := httptest.NewRequest(http.MethodGet, "/", nil)
	reqWithHeader.Header.Add("Forwarded", "proto=https")

	tab := map[*RequestBuilder]*http.Request{
		Get("/"):     httptest.NewRequest(http.MethodGet, "/", nil),
		Post("/"):    httptest.NewRequest(http.MethodPost, "/", nil),
		Put("/"):     httptest.NewRequest(http.MethodPut, "/", nil),
		Delete("/"):  httptest.NewRequest(http.MethodDelete, "/", nil),
		Patch("/"):   httptest.NewRequest(http.MethodPatch, "/", nil),
		Head("/"):    httptest.NewRequest(http.MethodHead, "/", nil),
		Options("/"): httptest.NewRequest(http.MethodOptions, "/", nil),
		Trace("/"):   httptest.NewRequest(http.MethodTrace, "/", nil),

		Post("http://foo.bar/test/path"):               httptest.NewRequest(http.MethodPost, "http://foo.bar/test/path", nil),
		Post("/").AddQueryParam("foo", "bar"):          httptest.NewRequest(http.MethodPost, "/?foo=bar", nil),
		Post("/").Body(&body):                          httptest.NewRequest(http.MethodPost, "/", &body),
		Get("/").AddHeader("Forwarded", "proto=https"): reqWithHeader,
	}

	for in, exp := range tab {
		act := in.Request()

		if diff := deep.Equal(exp, act); diff != nil {
			t.Error(diff)
		}
	}
}
