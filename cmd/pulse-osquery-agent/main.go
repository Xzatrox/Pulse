package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rcourtman/pulse-go-rewrite/internal/osqueryagent"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var defaultSystemPatterns = []string{
	"systemd*", "kworker*", "kthread*", "migration*", "ksoftirqd*", "kswapd*",
	"kcompactd*", "watchdog*", "*daemon", "dbus*", "rsyslog*", "cron*",
	"atd*", "accounts-daemon*", "polkit*", "rcu_*", "khugepaged*",
	"ksmd*", "kdevtmpfs*", "netns*", "kauditd*", "khungtaskd*",
	"oom_reaper*", "writeback*", "kblockd*", "ata_*", "scsi_*",
	"md*", "edac-*", "devfreq_*", "ksgxd*", "charger_manager*",
	"*-wq-*", "irq/*", "aio/*", "crypto*", "kintegrityd*",
	"bioset*", "kworker/*", "mm_percpu_wq*", "rcu_tasks_*",
	"idle_inject/*", "cpuhp/*", "ksoft*", "ktimersoftd/*",
	"getty*", "agetty*", "login*", "*udevd*", "lvmetad*",
	"multipathd*", "iscsid*", "rpcbind*", "rpc.*", "nfsd*",
	"lockd*", "rpciod*", "xprtiod*", "nfs*", "jbd2/*",
	"ext4-*", "xfs-*", "btrfs-*", "dm_*", "raid*",
	"md*_*", "loop*", "usb-storage*", "scsi_eh_*", "scsi_tmf_*",
	"iscsi_*", "fc_*", "nvme*", "mmc*", "spi*",
	"i2c*", "hci*", "bluetooth*", "cfg80211*", "wpa_supplicant*",
	"dhclient*", "NetworkManager*", "ModemManager*", "avahi*",
	"cups*", "acpid*", "thermald*", "irqbalance*", "mcelog*",
	"smartd*", "lvm*", "mdadm*", "auditd*", "audispd*",
	"sedispatch*", "abrtd*", "abrt-*", "rtkit*", "udisksd*",
	"upowerd*", "packagekitd*", "colord*", "geoclue*",
}

func main() {
	// Flags
	serverURL := flag.String("url", "http://localhost:7655", "Pulse server URL")
	apiToken := flag.String("api-token", "", "API token for authentication")
	agentID := flag.String("agent-id", "", "Agent ID (defaults to hostname)")
	interval := flag.Duration("interval", 60*time.Second, "Collection interval")
	filterMode := flag.String("filter-mode", "none", "Filter mode: none, basic, aggressive (default: none)")
	excludePatterns := flag.String("exclude-patterns", "", "Comma-separated patterns to exclude (supports wildcards: systemd*,*worker*,kthread)")
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

	// Parse exclude patterns
	var patterns []string
	
	// Apply filter mode presets
	switch *filterMode {
	case "basic":
		patterns = []string{
			"systemd*", "kworker*", "kthread*", "migration*", "ksoftirqd*",
			"kswapd*", "kcompactd*", "watchdog*", "*daemon", "dbus*",
		}
	case "aggressive":
		patterns = defaultSystemPatterns
	case "none":
		// No preset patterns
	default:
		log.Warn().Str("mode", *filterMode).Msg("Unknown filter mode, using 'none'")
	}
	
	// Add custom patterns
	if *excludePatterns != "" {
		customPatterns := strings.Split(*excludePatterns, ",")
		for _, p := range customPatterns {
			patterns = append(patterns, strings.TrimSpace(p))
		}
	}

	log.Info().
		Str("agent_id", *agentID).
		Str("server_url", *serverURL).
		Str("filter_mode", *filterMode).
		Int("exclude_pattern_count", len(patterns)).
		Dur("interval", *interval).
		Msg("Starting osquery agent")

	// Create agent
	agent, err := osqueryagent.New(osqueryagent.Config{
		AgentID:         *agentID,
		PulseURL:        *serverURL,
		APIToken:        *apiToken,
		Interval:        *interval,
		ExcludePatterns: patterns,
		Logger:          &log.Logger,
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
