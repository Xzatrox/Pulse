package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rcourtman/pulse-go-rewrite/internal/osqueryagent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// Flags
	serverURL := flag.String("url", "http://localhost:7655", "Pulse server URL")
	agentID := flag.String("agent-id", "", "Agent ID (defaults to hostname)")
	interval := flag.Duration("interval", 60*time.Second, "Collection interval")
	debug := flag.Bool("debug", false, "Enable debug logging")
	flag.Parse()

	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if *debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// Get agent ID
	if *agentID == "" {
		hostname, err := os.Hostname()
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to get hostname")
		}
		*agentID = hostname
	}

	log.Info().
		Str("agent_id", *agentID).
		Str("server_url", *serverURL).
		Dur("interval", *interval).
		Msg("Starting osquery agent")

	// Create agent
	agent, err := osqueryagent.New(osqueryagent.Config{
		AgentID:  *agentID,
		PulseURL: *serverURL,
		Interval: *interval,
		Logger:   &log.Logger,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create osquery agent")
	}

	// Handle shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Run agent
	if err := agent.Run(ctx); err != nil {
		log.Fatal().Err(err).Msg("Agent failed")
	}

	log.Info().Msg("Agent stopped")
}
