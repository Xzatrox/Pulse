package osqueryagent

import (
	"testing"

	"github.com/rs/zerolog"
)

func TestFilterProcesses(t *testing.T) {
	logger := zerolog.Nop()
	
	tests := []struct {
		name     string
		patterns []string
		input    []Process
		expected int
	}{
		{
			name:     "no patterns - no filtering",
			patterns: []string{},
			input: []Process{
				{Name: "systemd"},
				{Name: "chrome"},
				{Name: "kworker"},
			},
			expected: 3,
		},
		{
			name:     "exact match",
			patterns: []string{"systemd"},
			input: []Process{
				{Name: "systemd"},
				{Name: "chrome"},
			},
			expected: 1,
		},
		{
			name:     "prefix wildcard",
			patterns: []string{"systemd*"},
			input: []Process{
				{Name: "systemd"},
				{Name: "systemd-resolved"},
				{Name: "chrome"},
			},
			expected: 1,
		},
		{
			name:     "suffix wildcard",
			patterns: []string{"*worker"},
			input: []Process{
				{Name: "kworker"},
				{Name: "chrome"},
			},
			expected: 1,
		},
		{
			name:     "contains wildcard",
			patterns: []string{"*work*"},
			input: []Process{
				{Name: "kworker"},
				{Name: "networkd"},
				{Name: "chrome"},
			},
			expected: 1,
		},
		{
			name:     "multiple patterns",
			patterns: []string{"systemd*", "kworker", "*daemon"},
			input: []Process{
				{Name: "systemd"},
				{Name: "systemd-resolved"},
				{Name: "kworker"},
				{Name: "rsyslogd"},
				{Name: "chrome"},
			},
			expected: 1,
		},
		{
			name:     "case insensitive",
			patterns: []string{"SYSTEMD*"},
			input: []Process{
				{Name: "systemd"},
				{Name: "Systemd-resolved"},
				{Name: "chrome"},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			agent := &Agent{
				cfg: Config{
					ExcludePatterns: tt.patterns,
					Logger:          &logger,
				},
			}
			
			result := agent.filterProcesses(tt.input)
			if len(result) != tt.expected {
				t.Errorf("expected %d processes, got %d", tt.expected, len(result))
			}
		})
	}
}

func TestFilterServices(t *testing.T) {
	logger := zerolog.Nop()
	
	agent := &Agent{
		cfg: Config{
			ExcludePatterns: []string{"systemd*", "*daemon"},
			Logger:          &logger,
		},
	}
	
	input := []Service{
		{Name: "systemd-resolved.service"},
		{Name: "rsyslogd.service"},
		{Name: "ssh.service"},
		{Name: "cron.service"},
	}
	
	result := agent.filterServices(input)
	if len(result) != 2 {
		t.Errorf("expected 2 services, got %d", len(result))
	}
}
