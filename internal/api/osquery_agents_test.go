package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestOsqueryAgentHandlers_HandleReport(t *testing.T) {
	tmpDir := t.TempDir()
	h := NewOsqueryAgentHandlers(tmpDir)

	report := map[string]interface{}{
		"processes": []map[string]interface{}{
			{"pid": "1234", "name": "test", "path": "/usr/bin/test"},
		},
		"services": []map[string]interface{}{
			{"name": "sshd", "state": "running", "status": "active"},
		},
		"timestamp": "2024-01-01T00:00:00Z",
	}

	body, _ := json.Marshal(report)
	req := httptest.NewRequest(http.MethodPost, "/api/agents/test-agent/osquery", bytes.NewReader(body))
	w := httptest.NewRecorder()

	h.HandleReport(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestOsqueryAgentHandlers_HandleAllReports(t *testing.T) {
	tmpDir := t.TempDir()
	h := NewOsqueryAgentHandlers(tmpDir)

	req := httptest.NewRequest(http.MethodGet, "/api/osquery/reports", nil)
	w := httptest.NewRecorder()

	h.HandleAllReports(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestExtractAgentID(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/api/agents/test-123/osquery", "test-123"},
		{"/api/agents/node-1/osquery", "node-1"},
		{"/api/invalid", "unknown"},
	}

	for _, tt := range tests {
		req := httptest.NewRequest(http.MethodGet, tt.path, nil)
		result := extractAgentID(req)
		if result != tt.expected {
			t.Errorf("path %s: expected %s, got %s", tt.path, tt.expected, result)
		}
	}
}
