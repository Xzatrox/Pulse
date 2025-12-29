package osqueryagent

import (
	"context"
	"os/exec"
	"time"

	"github.com/rs/zerolog"
)

type Config struct {
	PulseURL           string
	APIToken           string
	Interval           time.Duration
	AgentID            string
	InsecureSkipVerify bool
	Logger             *zerolog.Logger
	ExcludePatterns    []string
}

type Agent struct {
	cfg           Config
	osqueryBinary string
}

func New(cfg Config) (*Agent, error) {
	binary, err := exec.LookPath("osqueryi")
	if err != nil {
		return nil, err
	}
	return &Agent{cfg: cfg, osqueryBinary: binary}, nil
}

func (a *Agent) Run(ctx context.Context) error {
	ticker := time.NewTicker(a.cfg.Interval)
	defer ticker.Stop()

	a.cfg.Logger.Info().Str("binary", a.osqueryBinary).Msg("osquery agent running")

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			a.collect()
		}
	}
}

func (a *Agent) collect() {
	processes, err := a.collectProcesses()
	if err != nil {
		a.cfg.Logger.Warn().Err(err).Msg("Failed to collect processes")
		return
	}

	services, err := a.collectServices()
	if err != nil {
		a.cfg.Logger.Warn().Err(err).Msg("Failed to collect services")
		return
	}

	// Apply filters
	processes = a.filterProcesses(processes)
	services = a.filterServices(services)

	a.cfg.Logger.Debug().
		Int("processes", len(processes)).
		Int("services", len(services)).
		Msg("Collected osquery data")

	if err := a.sendReport(processes, services); err != nil {
		a.cfg.Logger.Warn().Err(err).Msg("Failed to send report")
	} else {
		a.cfg.Logger.Debug().Msg("Report sent successfully")
	}
}
