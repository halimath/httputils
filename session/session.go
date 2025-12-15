package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/halimath/httputils"
	"github.com/halimath/kvlog"
)

// Private type for the context key
type contextKeyType string

// Sentinel value used as the context key to hold the [Session].
const contextKey contextKeyType = "contextKey"

// FromRequest returns the session associated with r. This is equivalent to
//
//	FromContext(r.Context())
func FromRequest(r *http.Request) Session {
	return FromContext(r.Context())
}

// FromContext returns the Session associated with ctx. If it does not exist,
// a nil Session is returned.
func FromContext(ctx context.Context) Session {
	v := ctx.Value(contextKey)
	if v == nil {
		return nil
	}

	s, ok := v.(Session)
	if !ok {
		panic(fmt.Sprintf("weired non session value found in context: %v", s))
	}
	return s
}

// withSession wraps s inside a [context.Context] using ctx as its parent and
// returns this context.
func withSession(ctx context.Context, s Session) context.Context {
	return context.WithValue(ctx, contextKey, s)
}

// --

// Session defines the interface for all session implementations. The design of
// this interface supports different implementations that include lazy loading
// or remote storing.
type Session interface {
	// ID returns the session’s ID.
	ID() string

	// RenewID generates a new ID for this session and stores it. Use this method
	// to renew a sessions’s id after certain events, such as authentication or
	// priviledge changes, as an addition security measure.
	RenewID()

	// Get gets the value associated with key from the Session and returns it
	// if found. The key is not found nil is returned.
	Get(key string) any

	// Set sets the value associated with key to val.
	Set(key string, val any)

	// Delete deletes key from this session.
	Delete(key string)

	// LastAccessed returnes the time stamp this session has been last accessed.
	LastAccessed() time.Time

	// Updates the last accessed timestamp for this session.
	SetLastAccessed(time.Time)
}

// --

// Get is a generic convenience to get and convert a value from a [Session].
// If key is not found in s or if the value for key in s is not of type T,
// T’s default value is returned.
func Get[T any](s Session, key string) (t T) {
	val := s.Get(key)
	if val == nil {
		return
	}

	t, _ = val.(T)
	return
}

// --

var ErrSessionNotFound = errors.New("session not found")

// Store defines the interface for session backend storage. It’s the store’s
// responsibility to synchronize concurrent access accordingly.
type Store interface {
	// Create creates a new session, stores it in this store and returns it.
	Create() (ses Session, err error)

	// Get retrieves the session associated with id and returns it. If id is
	// not found, [ErrSessionNotFound] is returned. If an error occurs while
	// looking up the session, a non-nil error is returned.
	Load(id string) (Session, error)

	// Set sets the session for id to s. If id already exists its value gets
	// overwritten. It returns an error if the operation cannot be performed.
	Store(s Session) error
}

// --

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
// session id is stored in a HTTP cookie with the name set to __Secure-Session-ID,
// the path to /, max-age to 30min and SameSite set to
// strict. Secure is set to true if the request uses HTTPS. Use
// [WithCookieOptions] to customize the cookie. HttpOnly set always set to true.
//
// The middleware adds the [Session] associated with each request to the
// request’s context; use [FromContext] function to extract the session from
// this context.
func NewMiddleware(opts ...Option) httputils.Middleware {
	mw := &middleware{
		cookie: CookieOpts{
			Name:     "__Secure-Session-ID",
			Path:     "/",
			MaxAge:   30 * time.Minute,
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

				http.SetCookie(w, &http.Cookie{
					Name:     mw.cookie.Name,
					Value:    ses.ID(),
					Domain:   mw.cookie.Domain,
					HttpOnly: true,
					Path:     mw.cookie.Path,
					Secure:   r.URL.Scheme == "https",
					MaxAge:   int(mw.cookie.MaxAge.Seconds()),
					SameSite: mw.cookie.SameSite,
				})
			}

			ses.SetLastAccessed(time.Now())

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

// --

const sessionIDBytes = 32 // 32 bytes = 256 bits of entropy

// GenerateSessionID generates a random, cryptographically secure session id.
func GenerateSessionID() string {
	buf := make([]byte, sessionIDBytes)

	_, err := rand.Read(buf)
	if err != nil {
		panic(fmt.Sprintf("unable to generate session id: %v", err))
	}

	return base64.RawURLEncoding.EncodeToString(buf)
}
