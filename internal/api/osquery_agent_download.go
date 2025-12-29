package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

// handleDownloadOsqueryAgent serves the pulse-osquery-agent binary
func (r *Router) handleDownloadOsqueryAgent(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Prevent caching
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	platform := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("platform")))
	arch := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("arch")))

	if platform == "" {
		platform = "linux"
	}
	if arch == "" {
		arch = "amd64"
	}

	// Build binary name
	binaryName := "pulse-osquery-agent-" + platform + "-" + arch
	if platform == "windows" {
		binaryName += ".exe"
	}

	// Search paths
	searchPaths := []string{
		filepath.Join(pulseBinDir(), binaryName),
		filepath.Join("/opt/pulse", binaryName),
		filepath.Join("/app", binaryName),
		filepath.Join(r.projectRoot, "bin", binaryName),
	}

	for _, candidate := range searchPaths {
		if candidate == "" {
			continue
		}

		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}

		checksum, err := r.cachedSHA256(candidate, info)
		if err != nil {
			log.Error().Err(err).Str("path", candidate).Msg("Failed to compute osquery agent checksum")
			continue
		}

		file, err := os.Open(candidate)
		if err != nil {
			log.Error().Err(err).Str("path", candidate).Msg("Failed to open osquery agent binary")
			continue
		}

		w.Header().Set("X-Checksum-Sha256", checksum)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+binaryName+"\"")
		http.ServeContent(w, req, binaryName, info.ModTime(), file)
		file.Close()
		return
	}

	// Fallback: redirect to GitHub releases
	githubURL := "https://github.com/rcourtman/Pulse/releases/latest/download/" + binaryName
	log.Info().Str("binary", binaryName).Str("redirect", githubURL).Msg("Local osquery agent binary not found, redirecting to GitHub")
	w.Header().Set("X-Served-From", "github-redirect")
	http.Redirect(w, req, githubURL, http.StatusTemporaryRedirect)
}

// handleDownloadOsqueryAgentChecksum serves the SHA256 checksum for the osquery agent binary
func (r *Router) handleDownloadOsqueryAgentChecksum(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	platform := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("platform")))
	arch := strings.ToLower(strings.TrimSpace(req.URL.Query().Get("arch")))

	if platform == "" {
		platform = "linux"
	}
	if arch == "" {
		arch = "amd64"
	}

	binaryName := "pulse-osquery-agent-" + platform + "-" + arch
	if platform == "windows" {
		binaryName += ".exe"
	}

	searchPaths := []string{
		filepath.Join(pulseBinDir(), binaryName),
		filepath.Join("/opt/pulse", binaryName),
		filepath.Join("/app", binaryName),
		filepath.Join(r.projectRoot, "bin", binaryName),
	}

	for _, candidate := range searchPaths {
		if candidate == "" {
			continue
		}

		info, err := os.Stat(candidate)
		if err != nil || info.IsDir() {
			continue
		}

		checksum, err := r.cachedSHA256(candidate, info)
		if err != nil {
			log.Error().Err(err).Str("path", candidate).Msg("Failed to compute checksum")
			continue
		}

		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte(checksum + "  " + binaryName + "\n"))
		return
	}

	http.Error(w, "Checksum not found", http.StatusNotFound)
}

// handleDownloadOsqueryInstallScript serves the install-osquery-agent.sh script
func (r *Router) handleDownloadOsqueryInstallScript(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")

	scriptPath := "/opt/pulse/scripts/install-osquery-agent.sh"
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		scriptPath = filepath.Join(r.projectRoot, "scripts", "install-osquery-agent.sh")
		if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
			githubURL := "https://raw.githubusercontent.com/rcourtman/Pulse/main/scripts/install-osquery-agent.sh"
			log.Info().Msg("Local install-osquery-agent.sh not found, redirecting to GitHub")
			http.Redirect(w, req, githubURL, http.StatusTemporaryRedirect)
			return
		}
	}

	w.Header().Set("Content-Type", "text/x-shellscript")
	w.Header().Set("Content-Disposition", "inline; filename=\"install-osquery-agent.sh\"")
	http.ServeFile(w, req, scriptPath)
}
