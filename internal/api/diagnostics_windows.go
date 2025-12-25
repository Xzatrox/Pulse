//go:build windows
// +build windows

package api

import (
	"net/http"
)

func (r *Router) handleDiagnostics(w http.ResponseWriter, req *http.Request) {
	http.Error(w, "Diagnostics not supported on Windows", http.StatusNotImplemented)
}
