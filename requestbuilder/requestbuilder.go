// Package requestbuilder contains a builder using httptest.NewRequest
// with methods supporting invocation chaining.
package requestbuilder

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
)

type header struct {
	key, value string
}

// RequestBuilder implements a builder for http.Request values
// using httptest.NewRequest internally.
type RequestBuilder struct {
	method string
	target *url.URL
	body   io.Reader
	header []header
}

// Get creates a new builder using HTTP verb GET.
// target is used the same way as described for httptest.NewRequest.
func Get(target string) *RequestBuilder {
	return New(http.MethodGet, target)
}

// Get creates a new builder using HTTP verb POST.
// target is used the same way as described for httptest.NewRequest.
func Post(target string) *RequestBuilder {
	return New(http.MethodPost, target)
}

// Get creates a new builder using HTTP verb PUT.
// target is used the same way as described for httptest.NewRequest.
func Put(target string) *RequestBuilder {
	return New(http.MethodPut, target)
}

// Get creates a new builder using HTTP verb DELETE.
// target is used the same way as described for httptest.NewRequest.
func Delete(target string) *RequestBuilder {
	return New(http.MethodDelete, target)
}

// Get creates a new builder using HTTP verb PATCH.
// target is used the same way as described for httptest.NewRequest.
func Patch(target string) *RequestBuilder {
	return New(http.MethodPatch, target)
}

// Get creates a new builder using HTTP verb OPTIONS.
// target is used the same way as described for httptest.NewRequest.
func Options(target string) *RequestBuilder {
	return New(http.MethodOptions, target)
}

// Get creates a new builder using HTTP verb HEAD.
// target is used the same way as described for httptest.NewRequest.
func Head(target string) *RequestBuilder {
	return New(http.MethodHead, target)
}

// Get creates a new builder using HTTP verb TRACE.
// target is used the same way as described for httptest.NewRequest.
func Trace(target string) *RequestBuilder {
	return New(http.MethodTrace, target)
}

// New creates a new request builder with method and target.
// target is used the same way as described for httptest.NewRequest.
func New(method string, target string) *RequestBuilder {
	t, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	return &RequestBuilder{
		method: method,
		target: t,
		body:   nil,
	}
}

// Body sets the request's body and returns the builder.
func (r *RequestBuilder) Body(body io.Reader) *RequestBuilder {
	r.body = body
	return r
}

// AddQueryParam adds a URL query parameter and returns the builder.
func (r *RequestBuilder) AddQueryParam(key, value string) *RequestBuilder {
	q := r.target.Query()
	q.Add(key, value)
	r.target.RawQuery = q.Encode()
	return r
}

// AddHeader adds a request header and returns the builder.
func (r *RequestBuilder) AddHeader(key, value string) *RequestBuilder {
	r.header = append(r.header, header{key, value})
	return r
}

// Request creates a new http.Request using httptest.NewRequest internally.
// After the request has been created the builder may be used to build more
// requests including requests with modified values.
func (r *RequestBuilder) Request() *http.Request {
	req := httptest.NewRequest(r.method, r.target.String(), r.body)

	for _, h := range r.header {
		req.Header.Add(h.key, h.value)
	}

	return req
}
