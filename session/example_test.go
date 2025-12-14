package session_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"github.com/halimath/httputils/session"
)

func Example() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sessionMiddleware := session.NewMiddleware(
		session.WithStore(session.NewInMemoryStore(
			session.WithContext(ctx),
			session.WithMaxTTL(5*time.Minute),
		)),
	)

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ses := session.FromRequest(r)
		c := session.Get[int](ses, "req_count")
		fmt.Println(c)
		ses.Set("req_count", c+1)
		w.WriteHeader(http.StatusNoContent)
	})

	handler = sessionMiddleware(handler)

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, r)
	sessionCookie, err := http.ParseSetCookie(w.Header().Get("Set-Cookie"))
	if err != nil {
		panic(err)
	}

	r = httptest.NewRequest("GET", "/", nil)
	r.AddCookie(sessionCookie)
	w = httptest.NewRecorder()

	handler.ServeHTTP(w, r)

	// Output:
	// 0
	// 1
}
