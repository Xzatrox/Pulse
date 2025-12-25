//go:build test
// +build test

package api

import (
	"embed"
	"net/http"
)

//go:embed frontend_stub/index.html
var frontendFS embed.FS

func serveFrontendHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
