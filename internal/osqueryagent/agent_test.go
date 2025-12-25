package osqueryagent

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rs/zerolog"
)

func TestNew(t *testing.T) {
	logger := zerolog.Nop()
	cfg := Config{
		PulseURL: "http://localhost:7655",
		APIToken: "test-token",
		AgentID:  "test-agent",
		Interval: 30 * time.Second,
		Logger:   &logger,
	}

	agent, err := New(cfg)
	if err != nil {
		t.Skip("osqueryi not found in PATH, skipping test")
	}
	if agent == nil {
		t.Error("expected agent, got nil")
	}
}

func TestAgent_Run(t *testing.T) {
	logger := zerolog.Nop()
	cfg := Config{
		PulseURL: "http://localhost:7655",
		APIToken: "test-token",
		AgentID:  "test-agent",
		Interval: 100 * time.Millisecond,
		Logger:   &logger,
	}

	agent, err := New(cfg)
	if err != nil {
		t.Skip("osqueryi not found in PATH, skipping test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err = agent.Run(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("Run failed: %v", err)
	}
}

func TestAgent_CollectAndSend(t *testing.T) {
	reportReceived := false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost && r.URL.Path == "/api/agents/test-agent/osquery" {
			reportReceived = true
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	logger := zerolog.Nop()
	agent := &Agent{
		cfg: Config{
			PulseURL: server.URL,
			APIToken: "test-token",
			AgentID:  "test-agent",
			Logger:   &logger,
		},
		osqueryBinary: "/mock/osqueryi",
	}

	processes := []Process{{PID: "1", Name: "init", Path: "/sbin/init"}}
	services := []Service{{Name: "sshd", State: "running", Status: "active"}}

	err := agent.sendReport(processes, services)
	if err != nil {
		t.Errorf("sendReport failed: %v", err)
	}

	if !reportReceived {
		t.Error("report was not received by server")
	}
}
