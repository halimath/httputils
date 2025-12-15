package session

import (
	"context"
	"testing"

	"github.com/halimath/expect"
	"github.com/halimath/expect/is"
)

func TestFromContext(t *testing.T) {
	ctx := context.Background()
	s := NewInMemorySession()

	ctx = withSession(ctx, s)

	got := FromContext(ctx)
	expect.That(t,
		is.DeepEqualTo(got, s),
	)
}

func TestGet(t *testing.T) {
	s := NewInMemorySession()
	s.Set("int", 42)
	s.Set("string", "hello")

	i := Get[int](s, "int")
	expect.That(t,
		is.EqualTo(42, i),
	)

	str := Get[string](s, "string")
	expect.That(t,
		is.EqualTo("hello", str),
	)

	missing := Get[int](s, "nope")
	expect.That(t,
		is.EqualTo(0, missing),
	)
}

func TestInMemoryStore(t *testing.T) {
	t.Run("get", func(t *testing.T) {

		store := NewInMemoryStore()

		s, err := store.Load("missing")
		expect.That(t,
			is.Error(err, ErrSessionNotFound),
			is.DeepEqualTo(nil, s),
		)
	})

	t.Run("get_set", func(t *testing.T) {
		store := NewInMemoryStore()
		s := NewInMemorySession()
		s.Set("x", 1)

		err := store.Store(s)
		expect.That(t,
			is.NoError(err),
		)

		got, err := store.Load(s.ID())
		expect.That(t,
			is.NoError(err),
			is.DeepEqualTo(got, s),
		)
	})

	t.Run("get_set_renew_get", func(t *testing.T) {
		store := NewInMemoryStore()
		s := NewInMemorySession()

		err := store.Store(s)
		expect.That(t,
			is.NoError(err),
		)

		originalID := s.ID()
		got, err := store.Load(originalID)
		expect.That(t,
			is.NoError(err),
			is.DeepEqualTo(got, s),
		)

		s.RenewID()
		err = store.Store(s)
		expect.That(t,
			is.NoError(err),
		)

		_, err = store.Load(originalID)
		expect.That(t,
			is.Error(err, ErrSessionNotFound),
		)

		got, err = store.Load(s.ID())
		expect.That(t,
			is.NoError(err),
			is.DeepEqualTo(got, s),
		)
	})
}
