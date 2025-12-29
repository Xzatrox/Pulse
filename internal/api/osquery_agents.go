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
	store     OsqueryStoreInterface
	dataPath  string
}

func NewOsqueryAgentHandlers(dataPath string) *OsqueryAgentHandlers {
	store, err := NewOsqueryStore(dataPath)
	if err != nil {
		log.Error().Err(err).Str("dataPath", dataPath).Msg("Failed to initialize osquery store")
	}
	if store != nil {
		log.Info().Str("dataPath", dataPath).Msg("osquery store initialized")
	}
	return &OsqueryAgentHandlers{store: store, dataPath: dataPath}
}

func (h *OsqueryAgentHandlers) StartCleanupScheduler(retentionDays int) {
	if h.store == nil {
		log.Warn().Msg("Cannot start osquery cleanup scheduler - store not initialized")
		return
	}

	if retentionDays <= 0 {
		retentionDays = 7
	}

	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		runOsqueryCleanup(h.store, retentionDays)

		for range ticker.C {
			runOsqueryCleanup(h.store, retentionDays)
		}
	}()
}

func runOsqueryCleanup(store OsqueryStoreInterface, retentionDays int) {
	if osqueryStore, ok := store.(*OsqueryStore); ok {
		deleted, err := osqueryStore.CleanupOldReports(retentionDays)
		if err != nil {
			log.Warn().Err(err).Msg("Failed to cleanup osquery reports")
		} else if deleted > 0 {
			log.Info().Int64("deleted", deleted).Int("retentionDays", retentionDays).Msg("Cleaned up old osquery reports")
		}
	}
}

func (h *OsqueryAgentHandlers) ensureStore() error {
	if h.store != nil {
		return nil
	}
	store, err := NewOsqueryStore(h.dataPath)
	if err != nil {
		return err
	}
	h.store = store
	log.Info().Str("dataPath", h.dataPath).Msg("osquery store initialized (retry)")
	return nil
}

func (h *OsqueryAgentHandlers) HandleRegister(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var regReq struct {
		AgentID  string `json:"agent_id"`
		Hostname string `json:"hostname"`
		Version  string `json:"version"`
	}

	if err := json.NewDecoder(req.Body).Decode(&regReq); err != nil {
		log.Warn().Err(err).Msg("Failed to decode osquery registration request")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	agentID := regReq.AgentID
	if agentID == "" {
		agentID = regReq.Hostname
	}

	log.Info().
		Str("agent_id", agentID).
		Str("hostname", regReq.Hostname).
		Str("version", regReq.Version).
		Msg("osquery agent registered")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Agent registered successfully",
	})
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
		AgentID   string `json:"agent_id"`
		Processes []struct {
			PID         string   `json:"pid"`
			Name        string   `json:"name"`
			Path        string   `json:"path"`
			LogFiles    []string `json:"log_files"`
			LogCommand  string   `json:"log_command"`
			MemoryBytes string   `json:"memory_bytes"`
			Status      string   `json:"status"`
		} `json:"processes"`
		Services []struct {
			Name   string `json:"name"`
			State  string `json:"state"`
			Status string `json:"status"`
			Health string `json:"health"`
		} `json:"services"`
		Timestamp string `json:"timestamp"`
	}

	if err := json.NewDecoder(req.Body).Decode(&report); err != nil {
		log.Warn().Err(err).Msg("Failed to decode osquery report")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	agentID := report.AgentID
	if agentID == "" {
		agentID = extractAgentID(req)
	}
	log.Info().
		Str("agent_id", agentID).
		Int("processes", len(report.Processes)).
		Int("services", len(report.Services)).
		Msg("Received osquery report")

	if err := h.ensureStore(); err != nil {
		log.Error().Err(err).Msg("Failed to initialize osquery store")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Store not initialized",
		})
		return
	}

	timestamp, _ := time.Parse(time.RFC3339, report.Timestamp)
	if timestamp.IsZero() {
		timestamp = time.Now()
	}
	if err := h.store.SaveReport(agentID, report.Processes, report.Services, timestamp); err != nil {
		log.Error().Err(err).Str("agent_id", agentID).Msg("Failed to save osquery report")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to save report",
		})
		return
	}

	log.Debug().Str("agent_id", agentID).Int("processes", len(report.Processes)).Int("services", len(report.Services)).Msg("osquery report saved successfully")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Report received",
	})
}

func (h *OsqueryAgentHandlers) HandleAllReports(w http.ResponseWriter, req *http.Request) {
	if err := h.ensureStore(); err != nil {
		log.Error().Err(err).Msg("Failed to initialize osquery store")
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
	if err := h.ensureStore(); err != nil {
		log.Error().Err(err).Msg("Failed to initialize osquery store")
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
