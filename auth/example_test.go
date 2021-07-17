package auth_test

import (
	"encoding/base64"
	"net/http"
	"strings"

	"github.com/halimath/httputils/auth"
)

func Example() {
	// h is a http.Handler, that actually handles the request.
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")

		// We can assume here that auth is always set. See below
		a := auth.GetAuthorization(r.Context())

		switch a.(type) {
		case *auth.UsernamePassword:
			// Use username/password to authorize the usert
		case *auth.BearerToken:
			// Decode token and authorizes
		}
	})

	http.ListenAndServe(":1234",
		auth.Basic(
			auth.Bearer(
				auth.Authorized(h,
					auth.AuthenticationChallenge{
						Scheme: auth.AuthorizationSchemeBasic,
						Realm:  "test",
					},
					auth.AuthenticationChallenge{
						Scheme: auth.AuthorizationSchemeBearer,
						Realm:  "test",
					},
				),
			),
		),
	)
}

func Example_custom() {
	type HMAC struct {
		Username string
		MAC      []byte
	}

	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// ...
	})

	http.ListenAndServe(":1234",
		auth.AuthHandler(
			auth.Authorized(h,
				auth.AuthenticationChallenge{
					Scheme: auth.AuthorizationSchemeBasic,
					Realm:  "test",
				},
				auth.AuthenticationChallenge{
					Scheme: auth.AuthorizationSchemeBearer,
					Realm:  "test",
				},
			),
			"Hmac",
			func(credentials string) auth.Authorization {
				parts := strings.Split(credentials, ":")
				if len(parts) != 2 {
					return nil
				}

				mac, err := base64.StdEncoding.DecodeString(parts[1])
				if err != nil {
					return nil
				}

				return &HMAC{
					Username: parts[0],
					MAC:      mac,
				}
			},
		),
	)
}
