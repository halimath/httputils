package session

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"time"
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
