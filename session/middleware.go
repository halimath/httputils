package session

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/halimath/httputils"
	"github.com/halimath/kvlog"
)

type CookieOpts struct {
	Name     string
	Path     string
	Domain   string
	MaxAge   time.Duration
	SameSite http.SameSite
}

type middleware struct {
	store  Store
	cookie CookieOpts
}

// Option defines a mutator type to configure a middleware.
type Option func(*middleware)

// WithStore is an [Options] that configures the [Store] to use.
func WithStore(s Store) Option {
	return func(m *middleware) {
		m.store = s
	}
}

// WithCookieOptions is an [Option] that customizes the session cookie.
func WithCookieOptions(opts CookieOpts) Option {
	return func(m *middleware) {
		if opts.Name != "" {
			m.cookie.Name = opts.Name
		}
		if opts.Path != "" {
			m.cookie.Path = opts.Path
		}
		if opts.Domain != "" {
			m.cookie.Domain = opts.Domain
		}
		if opts.MaxAge != 0 {
			m.cookie.MaxAge = opts.MaxAge
		}
		if opts.SameSite != 0 {
			m.cookie.SameSite = opts.SameSite
		}
	}
}

// NewMiddleware creates a new HTTP middleware that adds session
// management. By default, the [Store] in use is an in-memory store. The
// session id is stored in a HTTP cookie with the name set to session_id,
// the path to /, max-age to 5min and SameSite set to
// strict. Secure is set to true if the request uses HTTPS. Use
// [WithCookieOptions] to customize the cookie. HttpOnly is always set to true.
// The cookie will automatically be prolonged on every request.
//
// The middleware adds the [Session] associated with each request to the
// request’s context; use [FromContext] function to extract the session from
// this context.
func NewMiddleware(opts ...Option) httputils.Middleware {
	mw := &middleware{
		cookie: CookieOpts{
			Name:     "session_id",
			Path:     "/",
			MaxAge:   5 * time.Minute,
			SameSite: http.SameSiteStrictMode,
		},
	}

	for _, opt := range opts {
		opt(mw)
	}

	if mw.store == nil {
		mw.store = NewInMemoryStore()
	}

	return func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			logger := kvlog.FromContext(r.Context())

			var ses Session
			var err error

			cookie, err := r.Cookie(mw.cookie.Name)
			if err != http.ErrNoCookie {
				id := cookie.Value
				ses, err = mw.store.Load(id)
				if err != nil {
					if !errors.Is(err, ErrSessionNotFound) {
						logger.Logs("failed to load session from store", kvlog.WithKV("id", id), kvlog.WithErr(err))

						http.Error(w, "internal server error", http.StatusInternalServerError)
						return
					}
					logger.Logs("Session with id not found", kvlog.WithKV("id", id))
				}
			}

			if ses == nil {
				ses, err = mw.store.Create()
				if err != nil {
					logger.Logs("failed to create new session", kvlog.WithErr(err))

					http.Error(w, "internal server error", http.StatusInternalServerError)
					return
				}
				logger.Logs("no previous session found; creating new one", kvlog.WithKV("id", ses.ID()))

			}

			ses.SetLastAccessed(time.Now())

			// Set cookie for every request to extend the cookie’s max age.
			http.SetCookie(w, &http.Cookie{
				Name:     mw.cookie.Name,
				Value:    ses.ID(),
				Domain:   mw.cookie.Domain,
				HttpOnly: true,
				Path:     mw.cookie.Path,
				Secure:   isSecureRequest(r),
				MaxAge:   int(mw.cookie.MaxAge.Seconds()),
				SameSite: mw.cookie.SameSite,
			})

			ctx := r.Context()
			r = r.WithContext(withSession(ctx, ses))

			handler.ServeHTTP(w, r)

			err = mw.store.Store(ses)
			if err != nil {
				// The response has already been commenced and we cannot send an error,
				// so we just log the error
				logger.Logs("failed to store session from store", kvlog.WithKV("id", ses.ID()), kvlog.WithErr(err))
			}
		})
	}
}

func isSecureRequest(r *http.Request) bool {
	// Direct TLS connection
	if r.TLS != nil {
		return true
	}

	// Forwarded request header
	if forwarded := r.Header.Get("Forwarded"); forwarded != "" {
		if strings.Contains(forwarded, "proto=https") {
			return true
		}
	}

	// Legacy X-Forwarded-Proto header
	if proto := r.Header.Get("X-Forwarded-Proto"); proto == "https" {
		return true
	}

	// Parsed request URL
	return r.URL.Scheme == "https"
}
