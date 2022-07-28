package cors_test

import (
	"net/http"

	"github.com/halimath/httputils/cors"
)

func Example_allOrigins() {
	// restAPI is a http.Handler that defines some kind of resource.
	restAPI := http.NewServeMux()

	http.ListenAndServe(":1234", cors.Middleware(restAPI))
}

func Example_customOrigins() {
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
}
