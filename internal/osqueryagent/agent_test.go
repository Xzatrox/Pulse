package osqueryagent

import (
	"context"
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
		t.Errorf("New failed: %v", err)
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

	agent, _ := New(cfg)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := agent.Run(ctx)
	if err != nil && err != context.DeadlineExceeded {
		t.Errorf("Run failed: %v", err)
	}
}
