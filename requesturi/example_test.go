package requesturi_test

import (
	"fmt"
	"net/http"

	"github.com/halimath/httputils/requesturi"
)

func Example() {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fmt.Sprintf("%s://%s/some/path", r.URL.Scheme, r.URL.Host), http.StatusTemporaryRedirect)
	})

	http.ListenAndServe(":1234", requesturi.Middleware(h, requesturi.Forwarded, requesturi.XForwarded))
}
