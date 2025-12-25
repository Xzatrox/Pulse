package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
)

type OsqueryStoreInterface interface {
	SaveReport(agentID string, processes, services interface{}, timestamp time.Time) error
	GetLatestReport(agentID string) (map[string]interface{}, error)
	GetAllLatestReports() (map[string]interface{}, error)
	Close() error
}

type OsqueryAgentHandlers struct {
	store OsqueryStoreInterface
}

func NewOsqueryAgentHandlers(dataPath string) *OsqueryAgentHandlers {
	store, err := NewOsqueryStore(dataPath)
	if err != nil {
		log.Error().Err(err).Msg("Failed to initialize osquery store")
		return &OsqueryAgentHandlers{}
	}
	return &OsqueryAgentHandlers{store: store}
}

func (h *OsqueryAgentHandlers) HandleReport(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodGet {
		h.handleGetReport(w, req)
		return
	}

	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var report struct {
		Processes []struct {
			PID      string   `json:"pid"`
			Name     string   `json:"name"`
			Path     string   `json:"path"`
			LogFiles []string `json:"log_files"`
		} `json:"processes"`
		Services []struct {
			Name   string `json:"name"`
			State  string `json:"state"`
			Status string `json:"status"`
		} `json:"services"`
		Timestamp string `json:"timestamp"`
	}

	if err := json.NewDecoder(req.Body).Decode(&report); err != nil {
		log.Warn().Err(err).Msg("Failed to decode osquery report")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	agentID := extractAgentID(req)
	log.Info().
		Str("agent_id", agentID).
		Int("processes", len(report.Processes)).
		Int("services", len(report.Services)).
		Msg("Received osquery report")

	if h.store != nil {
		timestamp, _ := time.Parse(time.RFC3339, report.Timestamp)
		if timestamp.IsZero() {
			timestamp = time.Now()
		}
		if err := h.store.SaveReport(agentID, report.Processes, report.Services, timestamp); err != nil {
			log.Error().Err(err).Msg("Failed to save osquery report")
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Report received",
	})
}

func (h *OsqueryAgentHandlers) HandleAllReports(w http.ResponseWriter, req *http.Request) {
	if h.store == nil {
		http.Error(w, "Store not initialized", http.StatusServiceUnavailable)
		return
	}

	reports, err := h.store.GetAllLatestReports()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get all osquery reports")
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

func (h *OsqueryAgentHandlers) handleGetReport(w http.ResponseWriter, req *http.Request) {
	if h.store == nil {
		http.Error(w, "Store not initialized", http.StatusServiceUnavailable)
		return
	}

	agentID := extractAgentID(req)
	report, err := h.store.GetLatestReport(agentID)
	if err != nil {
		log.Warn().Err(err).Str("agent_id", agentID).Msg("No osquery data found")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"processes": []interface{}{},
			"services":  []interface{}{},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(report)
}

func extractAgentID(req *http.Request) string {
	parts := strings.Split(strings.Trim(req.URL.Path, "/"), "/")
	for i, part := range parts {
		if part == "agents" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return "unknown"
}
