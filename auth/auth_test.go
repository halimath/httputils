package auth

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-test/deep"
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
			act := GetAuthorization(r.Context())
			if diff := deep.Equal(exp, act); diff != nil {
				t.Error(diff)
			}
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
			act := GetAuthorization(r.Context())
			if diff := deep.Equal(exp, act); diff != nil {
				t.Error(diff)
			}
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

		if w.Result().StatusCode != exp {
			t.Errorf("%v: expected %d but got %d", in, exp, w.Result().StatusCode)
		}
	}
}
