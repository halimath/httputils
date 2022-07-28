package auth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halimath/expect-go"
	"github.com/halimath/httputils/requestbuilder"
)

func TestBasicAuth(t *testing.T) {
	tab := map[*http.Request]Authorization{
		requestbuilder.Get("/").Request(): nil,

		requestbuilder.Get("/").AddHeader(HeaderAuthorization, "foo bar").Request(): nil,

		requestbuilder.Get("/").AddHeader(HeaderAuthorization, "Basic bar").Request(): nil,

		requestbuilder.Get("/").AddHeader(HeaderAuthorization, "Basic "+base64.StdEncoding.EncodeToString([]byte("foo"))).Request(): nil,

		requestbuilder.Get("/").AddHeader(HeaderAuthorization, "Basic dGVzdDoxMjPCow==").Request(): &UsernamePassword{
			Username: "test",
			Password: "123\u00A3",
		},
	}

	for in, exp := range tab {
		var w httptest.ResponseRecorder
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := GetAuthorization(r.Context())
			expect.That(t, got).Is(expect.DeepEqual(exp))
		})
		Basic(h).ServeHTTP(&w, in)
	}
}

func TestBearer(t *testing.T) {
	tab := map[*http.Request]Authorization{
		requestbuilder.Get("/").Request():                                                                                           nil,
		requestbuilder.Get("/").AddHeader(HeaderAuthorization, "foo bar").Request():                                                 nil,
		requestbuilder.Get("/").AddHeader(HeaderAuthorization, "Basic bar").Request():                                               nil,
		requestbuilder.Get("/").AddHeader(HeaderAuthorization, "Basic "+base64.StdEncoding.EncodeToString([]byte("foo"))).Request(): nil,
		requestbuilder.Get("/").AddHeader(HeaderAuthorization, "Bearer foobar").Request():                                           &BearerToken{Token: "foobar"},
	}

	for in, exp := range tab {
		var w httptest.ResponseRecorder
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := GetAuthorization(r.Context())
			expect.That(t, got).Is(expect.DeepEqual(exp))
		})
		Bearer(h).ServeHTTP(&w, in)
	}
}

func TestAuthorized(t *testing.T) {
	tab := map[*http.Request]int{
		requestbuilder.Get("/").Request(): http.StatusUnauthorized,
		requestbuilder.Get("/").AddHeader(HeaderAuthorization, "Basic "+base64.StdEncoding.EncodeToString([]byte("foo"))).Request(): http.StatusUnauthorized,
		requestbuilder.Get("/").AddHeader(HeaderAuthorization, "Bearer foobar").Request():                                           http.StatusOK,
	}

	h := Bearer(Authorized(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
		AuthenticationChallenge{
			Scheme: AuthorizationSchemeBasic,
			Realm:  "test",
		},
	))

	for in, exp := range tab {
		var w httptest.ResponseRecorder
		h.ServeHTTP(&w, in)

		expect.That(t, w.Result().StatusCode).Is(expect.Equal(exp))
	}
}
