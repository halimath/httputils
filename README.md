# httputils

Utilities for implementing HTTP services in [Go](https://golang.org)

![CI Status][ci-img-url] [![Go Report Card][go-report-card-img-url]][go-report-card-url] [![Package Doc][package-doc-img-url]][package-doc-url] [![Releases][release-img-url]][release-url]

This repo contains a library for the Golang programming language that provides utilities for implementing
HTTP services. The library focusses on using the standard library only and thus be agnostic about any 
other web framework.

# Installation

Use `go get` to install the libary with your project. You need Go >= 1.18 to use the lib. 

```
$ go get github.com/halimath/httputils
```

# Usage

`httputils` contains a set of different features that can be used independently or together. The following
sections each describe a single feature.

## Authorization

`httputils` contains a HTTP middleware that handles HTTP Authorization. The middleware extracts the
authorization credentials and stores them in the request's context before forwarding the request to
the next handler.

Currently, _Basic Auth_ and _Bearer Token_ are supported but the middleware allows for an easy extension.

The following example demonstrates how to use the `auth` package.

```go
// h is a http.Handler, that actualy handles the request.
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
    auth.Bearer(
        auth.Basic(
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
```

In the example above `h` is a simple `http.Handler`; replace it with your "real" handler implementation. 
You can also use any kind of framework here as long as the framework's router implements the 
`http.Handler` interface.

The call to `ListenAndServe` uses three middleware that wrap each other with the inner most wrapping `h`. 
Let's go through them from outer to inner:

* `auth.Bearer` creates a middleware that tries to extract a _Token Bearer Authorization_ credentials
  from the request and - if found - stores the credentials in the requests's context. It always invokes
  the wrapped handler.
* `auth.Basic` creates a middleware that extracts any _Basic Auth_ credentials and stores them in the 
  requests context. It always invokes the wrapped handler.
* `auth.Authorized` creates a middleware that checks if the request's context contains a non-`nil`
  Authorization (extracted from either of the above middlewares). If such an authorization is found
  the wrapped handler is invoked. If no authoriation has been found, the request is rejected with
  a HTTP status code `401 Unauthorized` and a `WWW-Authentication` header is added with the given
  HTTP authentication challenges.

It's important to keep the order of the middlewares correct: 

* If you put `auth.Authorized` first then _every request_ will be rejected as there are no handlers 
storing an `Authorization` value in the context.
* The order of `Bearer` and `Basic` is important only for requests that contain _both_ authorizations
  (i.e. by sending two `Authorization` header). The last (successful) middleware overwrites any 
  Authorization value stored by a middleware that ran previously.

### How to implement your own Authorization scheme

HTTP Authorization is pretty flexible so chances are that you need a custom implementation to grab the
user's credentials from a request. If you want to use the `Authorized` middleware you need to do the 
following:

* Create a type holding the user's credentials. This type implements `auth.Authorization` which is an
  empty interface.
* Create a middleware that extracts the credentials from a request and calls `auth.WithAuthorization` to
  create a new context holding the credentials. If your implementation also uses the HTTP `Authorization`
  header with a custom scheme, you may use `auth.AuthHandler` to simplify the implementation by providing
  a function that creates an `Authorization` value from a credentials string.

Here is a sketched example that demonstrates how to build some kind of HMAC authorization. The idea is, 
that requests carry an `Authorization`-header with a scheme `Hmac` that contains a _keyed hashed method
authentication code_ for the request's URL signed with a user's secret. Username and hmac are separated
with a single colon; the HMAC is base64-encoded, such as

```
GET /foo/bar HTTP/1.1
Authorization: Hmac john.doe:eLKW1g44EJ52qiF7kFbzma7zf61yE0x8gUO2daRwqss=
```

The example uses a SHA256 HMAC with the key `secret`. You can calculate it with 

```
echo -n "/foo/bar" | openssl dgst -sha256 -hmac "secret" -binary | openssl enc -base64 -A
```

The following code demonstrates how to set up an authorization handler implementing the above. Note that
the middleware does not verify the HMAC - it only performs the Base64 decoding.

```go
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
```

### A note on how to verify the credentials

You may have noted that none of the above middlewares that extract user credentials actually performs a
verificates besides some syntax checking. This task is intentionally left off the framework. The reason
for that is that the decision where to do authorization is a highly opinionated question with
different people argumenting for different directions. While some which to perform this step as part of
the response handling, others seek to implement this as part of the business layer (a _service_, domain 
function or whatever else is used to implement business logic). The `auth` package favors none of those
opinions and allows both to be implemented with ease. 

If you want to do the verification as part of the request handling, simply create another middleware
positioned after the `Authorized` middleware that does the verification. If you want to implement the
verification in a different software layer, simply pass the request's context to the business function
(which in modern Go is a generally good advice) and use `auth.GetAuthorization` to read the credentials.

## Request URI

The `requesturi` package contains a HTTP middleware that augments some of the request's `URL` fields that
are left blank by default. The resulting `URL` can be used to reconstruct the requested URI _as specified
by the client_. This is very usefull when creating dynanic links, redirect URLs or OAuth return URLs from
what the user "sees". 

The package also provides functions that extend the behavior when running behind a reverse proxy that
sets HTTP header to forward the original request information.

The following example configures the middleware for use behind a reverse proxy and reads the HTTP standard
`Forwarded`-header as well as the _defacto standard_ `X-Forwarded-*`-headers:

```go
h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    http.Redirect(w, r, fmt.Sprintf("%s://%s/some/path", r.URL.Scheme, r.URL.Host), http.StatusTemporaryRedirect)
})

http.ListenAndServe(":1234", requesturi.Middleware(h, requesturi.Forwarded, requesturi.XForwarded))
```

## CORS

Package `cors` provides a configurable middleware to handle _Cross Origin Resource Sharing_ (CORS). The
middleware injects response headers and handles pre-flight requests completely.

To generally allow access to all resources (i.e. endpoints) from all origins use something like this:

```go
// restAPI is a http.Handler that defines some kind of resource.
restAPI := http.NewServeMux()

http.ListenAndServe(":1234", cors.Middleware(restAPI))
```

To enable CORS for specific endpoints and/or origins, you can pass additional configuration arguments to the
middleware:

```go
// restAPI is a http.Handler that defines some kind of resource.
restAPI := http.NewServeMux()

http.ListenAndServe(":1234",
    cors.Middleware(
        restAPI,
        cors.Endpoint{
            Path: "/api/v1/resource1",
        },
        cors.Endpoint{
            Path:             "/api/v1/resource2",
            AllowMethods:     []string{http.MethodPost},
            AllowCredentials: true,
        },
    ),
)
```

## Request Builder (for tests)

Package `requestbuilder` contains a builder that can be used to build `http.Request` values during tests.
While package `httptest` provides a `NewRequest` function to create a request for tests, setting headers
requires you to use a local variable. The request builder allows you to set all kinds of request 
properties using methods that return the builder.

```go
accessToken := "..."
data, _ := os.Open("/some/file")

_ = requestbuilder.Post("https://example.com/path/to/resource").
    Body(data).
    AddHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
    Request()
```

This works extremely well when using 
[table driven tests](https://github.com/golang/go/wiki/TableDrivenTests). The following code is
from the [`auth` package's tests](./auth/auth_test.gol):

```go
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

```

## Buffered Response

Package `github.com/halimath/httputils/bufferedresponse` provides a type `ResponseWriter` that satisfies the interface
[`http.ResponseWriter`](https://pkg.go.dev/net/http#ResponseWriter) as an in-memory buffered implementation.
The type "collects" all headers, status code and body bytes written and can then "replay" the response on
any (even multiple) `http.ResponseWriter`s.

Use this buffer implementation when implementing middlewares or request handlers that need a way to "rewind"
the response and start over (i.e. for handling errors).

## Response

Package `github.com/halimath/httputils/reponse` provides several functions to easily create responses from
http handler methods. These functions are built on the `bufferedresponse` package and provide easy to use,
easy to extend builting of http responses.

See the package doc and the corresponding tests for examples.

### Problem JSON

One special response helper is capable of sending problem details as described in [RFC9457]. The Problem
Details RFC defines a JSON (and XML) structure as well as some rules on the field's semantics to report
useful details from problem results. This module only implements the JSON representation of the RFC.

[RFC9457]: https://www.rfc-editor.org/rfc/rfc9457

## `errmux`

Package `errmux` provides an augmented version of `http.ServeMux` which accept handler methods that return
`error` values. The multiplexer uses a `http.ServeMux` under the hood and supports all the patterns supported
by the Go version in use (i.e. all advanced patterns introduced with Go 1.22 if a version >= 1.22 is used).

Any error returned from a handler will be caught and the response written so far will be discarded. The error
is then handled by an error handler which may be customized producing a final result to send to the client.

See the following example for a short demonstration:

```go
mux := errmux.NewServeMux()

errMissingQueryParameter := errors.New("missing query parameter")

mux.HandleFunc("/echo", func(w http.ResponseWriter, r *http.Request) error {
    if msg := r.URL.Query().Get("msg"); len(msg) > 0 {
        return response.PlainText(w, r, msg)
    }

    return fmt.Errorf("%w: %s", errMissingQueryParameter, "msg")
})

mux.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
    if errors.Is(err, errMissingQueryParameter) {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    http.Error(w, err.Error(), http.StatusInternalServerError)
}

http.ListenAndServe(":8080", mux)
```

# License

Copyright 2021 - 2024 Alexander Metzner
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

[http://www.apache.org/licenses/LICENSE-2.0](http://www.apache.org/licenses/LICENSE-2.0)

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

[ci-img-url]: https://github.com/halimath/httputils/workflows/CI/badge.svg
[go-report-card-img-url]: https://goreportcard.com/badge/github.com/halimath/httputils
[go-report-card-url]: https://goreportcard.com/report/github.com/halimath/httputils
[package-doc-img-url]: https://img.shields.io/badge/GoDoc-Reference-blue.svg
[package-doc-url]: https://pkg.go.dev/github.com/halimath/httputils
[release-img-url]: https://img.shields.io/github/v/release/halimath/httputils.svg
[release-url]: https://github.com/halimath/httputils/releases
