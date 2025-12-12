package auth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/halimath/expect"
	"github.com/halimath/expect/is"
	"github.com/halimath/httputils"
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

	for in, want := range tab {
		var w httptest.ResponseRecorder
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := GetAuthorization(r.Context())
			expect.That(t, is.DeepEqualTo(got, want))
		})
		Basic()(h).ServeHTTP(&w, in)
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

	for in, want := range tab {
		var w httptest.ResponseRecorder
		h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := GetAuthorization(r.Context())
			expect.That(t, is.DeepEqualTo(got, want))
		})
		Bearer()(h).ServeHTTP(&w, in)
	}
}

func TestAuthorized(t *testing.T) {
	tab := map[*http.Request]int{
		requestbuilder.Get("/noAuthHeader").Request(): http.StatusUnauthorized,
		requestbuilder.Get("/basicAuthHeader").AddHeader(HeaderAuthorization, "Basic "+base64.StdEncoding.EncodeToString([]byte("foo"))).Request(): http.StatusUnauthorized,
		requestbuilder.Get("/bearerAuthHeader").AddHeader(HeaderAuthorization, "Bearer foobar").Request():                                          http.StatusOK,
	}

	h := httputils.Compose(
		Authorized(
			AuthenticationChallenge{
				Scheme: AuthorizationSchemeBasic,
				Realm:  "test",
			},
		),
		Bearer(),
	)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	for in, want := range tab {
		var w httptest.ResponseRecorder
		h.ServeHTTP(&w, in)

		expect.WithMessage(t, in.URL.Path).That(is.EqualTo(w.Result().StatusCode, want))
	}
}
