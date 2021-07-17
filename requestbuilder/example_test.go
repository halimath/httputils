package requestbuilder_test

import (
	"fmt"
	"os"

	"github.com/halimath/httputils/requestbuilder"
)

func Example() {
	accessToken := "..."
	data, _ := os.Open("/some/file")

	_ = requestbuilder.Post("https://example.com/path/to/resource").
		Body(data).
		AddHeader("Authorization", fmt.Sprintf("Bearer %s", accessToken)).
		Request()
}
