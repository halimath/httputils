// Package auth contains http middleware implementations handling the HTTP authorization.
// All middlewares only parse the provided authorization credentials and update the
// request's Context. See RFC 7235 for details on HTTP based authorization.
// (https://datatracker.ietf.org/doc/html/rfc7235)
package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
)

type contextKeyAuthType string

const contextKeyAuth contextKeyAuthType = "auth"

// WithAuthorization extends ctx with auth stored under a private key.
func WithAuthorization(ctx context.Context, auth Authorization) context.Context {
	return context.WithValue(ctx, contextKeyAuth, auth)
}

// GetAuthorization returns the Authorization stored in ctx or nil
// if no authorization are stored in ctx.
func GetAuthorization(ctx context.Context) Authorization {
	v := ctx.Value(contextKeyAuth)
	if v == nil {
		return nil
	}

	auth, ok := v.(Authorization)
	if !ok {
		return nil
	}

	return auth
}

const (
	// Name of the HTTP Authorization header as specified in RFC 7235, section 4.2
	// (https://datatracker.ietf.org/doc/html/rfc7235#section-4.2)
	HeaderAuthorization = "Authorization"

	// Authorization scheme used with basic authentication as specified in
	// RFC 7617, section 2
	// (https://datatracker.ietf.org/doc/html/rfc7617#section-2)
	AuthorizationSchemeBasic = "Basic"

	// Authorization scheme used with token bearer authorization as specified
	// in RFC 6750, section 2.1
	// (https://datatracker.ietf.org/doc/html/rfc6750#section-2.1)
	AuthorizationSchemeBearer = "Bearer"

	// WWW-Authenticate response header as specified in RFC 7235, section 4.1
	// (https://datatracker.ietf.org/doc/html/rfc7235#section-4.1)
	HeaderWWWAuthenticate = "WWW-Authenticate"
)

// AuthenticationChallenge implements a single authentication challenge
// returned with a HTTP status 401.
type AuthenticationChallenge struct {
	Scheme    string
	Realm     string
	UserProps map[string]string
}

func (a *AuthenticationChallenge) toHeader() string {
	var b strings.Builder
	b.WriteString(a.Scheme)
	fmt.Fprintf(&b, `realm="%s"`, strings.ReplaceAll(a.Realm, `"`, `\"`))

	for k, v := range a.UserProps {
		fmt.Fprintf(&b, `, %s="%s"`, strings.ReplaceAll(k, `"`, `\"`), strings.ReplaceAll(v, `"`, `\"`))
	}

	return b.String()
}

// Authorized creates a http middleware wrapping h that checks if the
// request carries an Authorization (using GetAuthorization). If no
// authorization is found, the request is rejected with a HTTP status
// 401 (Unauthorized). The response contains a WWW-Authenticate header
// with the given challanges.
func Authorized(h http.Handler, challenge AuthenticationChallenge, moreChallenges ...AuthenticationChallenge) http.Handler {
	var b strings.Builder
	b.WriteString(challenge.toHeader())

	for _, c := range moreChallenges {
		b.WriteString(", ")
		b.WriteString(c.toHeader())
	}

	wwwAuthenticateHeader := b.String()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if GetAuthorization(r.Context()) == nil {
			w.Header().Add(HeaderWWWAuthenticate, wwwAuthenticateHeader)
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	})
}

// Authorization is a tagging interface implemented by all types of authentication Authorization.
type Authorization interface{}

// --

// UsernamePassword implements an Authorization capturing Authorization provided
// via HTTP BasicAuth Auth. See RFC 7617.
// (https://datatracker.ietf.org/doc/html/rfc7617)
type UsernamePassword struct {
	Username string
	Password string
}

// Basic creates a http middleware wrapping h which extracts basic
// autorization credentials as specified in RFC 7617 and stores them
// in the request's context. Use GetAuthorization to extract the authorization.
// (https://datatracker.ietf.org/doc/html/rfc7617)
func Basic(h http.Handler) http.Handler {
	return AuthHandler(h, AuthorizationSchemeBasic, func(c string) Authorization {
		pair, err := base64.StdEncoding.DecodeString(c)
		if err != nil {
			return nil
		}

		namePassword := strings.Split(string(pair), ":")
		if len(namePassword) != 2 {
			return nil
		}

		return &UsernamePassword{
			Username: namePassword[0],
			Password: namePassword[1],
		}
	})
}

// --

// BearerToken implements Authorization capturing a bearer token
// as specified in RFC 6750.
// (https://datatracker.ietf.org/doc/html/rfc6750)
type BearerToken struct {
	Token string
}

// Bearer creates a http middleware wrapping h that performs
// token bearer authorization as specified in
// RFC 6750, section 2.1. Note that only header based authorization
// is implemented.
// (https://datatracker.ietf.org/doc/html/rfc6750#section-2.1)
func Bearer(h http.Handler) http.Handler {
	return AuthHandler(h, AuthorizationSchemeBearer, func(t string) Authorization {
		return &BearerToken{
			Token: t,
		}
	})
}

// AuthorizationBuilder builds an Authorization value from the given credentials string.
type AuthorizationBuilder func(credentials string) Authorization

// AuthHandler creates a http middleware wrapping h that accepts Authoriation request headers
// using the authorization scheme. It forwards the credentials given after scheme to ab in
// order to build an Authorization object.
func AuthHandler(h http.Handler, scheme string, ab AuthorizationBuilder) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auths, ok := r.Header[HeaderAuthorization]

		if ok {
			for _, auth := range auths {
				if !strings.HasPrefix(auth, scheme) {
					continue
				}

				a := ab(strings.TrimSpace(auth[len(scheme):]))

				if a != nil {
					r = r.WithContext(WithAuthorization(r.Context(), a))
				}
			}
		}

		h.ServeHTTP(w, r)
	})
}
