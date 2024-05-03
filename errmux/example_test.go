package errmux_test

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/halimath/httputils/errmux"
	"github.com/halimath/httputils/response"
)

func ExampleServeMux() {
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
}
