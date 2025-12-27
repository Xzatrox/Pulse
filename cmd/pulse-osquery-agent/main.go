package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rcourtman/pulse-go-rewrite/internal/osqueryagent"
	"github.com/rcourtman/pulse-go-rewrite/internal/utils"
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
	// Additional system services
	"postfix*", "master", "pickup", "qmgr", "smtp*", "sendmail*",
	"ssh.service", "sshd.service", "container-getty*", "console-getty*",
	"plymouth*", "modprobe@*", "dpkg-*", "display-manager*",
	"e2scrub*", "ldconfig*", "rescue.service", "apt-*",
	"rc-local*", "emergency.service", "initrd-*", "logrotate*",
	"kmod-*", "man-db*", "networking.service", "wtmpdb-*",
	"nftables*", "fstrim*", "connman*", "*-resolvconf*",
	"syslog.service", "postfix@*",
}

func main() {
	// Read environment variables
	envURL := utils.GetenvTrim("PULSE_URL")
	envToken := utils.GetenvTrim("PULSE_TOKEN")
	envAgentID := utils.GetenvTrim("PULSE_AGENT_ID")
	envInterval := utils.GetenvTrim("PULSE_INTERVAL")
	envFilterMode := utils.GetenvTrim("PULSE_OSQUERY_FILTER_MODE")
	envExcludePatterns := utils.GetenvTrim("PULSE_OSQUERY_EXCLUDE_PATTERNS")
	envLogLevel := utils.GetenvTrim("LOG_LEVEL")

	defaultInterval := 60 * time.Second
	if envInterval != "" {
		if parsed, err := time.ParseDuration(envInterval); err == nil {
			defaultInterval = parsed
		}
	}

	defaultFilterMode := "none"
	if envFilterMode != "" {
		defaultFilterMode = envFilterMode
	}

	defaultLogLevel := "info"
	if envLogLevel != "" {
		defaultLogLevel = envLogLevel
	}

	// Flags (env vars as defaults, flags override)
	configFile := flag.String("config", "", "Path to configuration file (YAML)")
	serverURL := flag.String("url", envURL, "Pulse server URL")
	apiToken := flag.String("api-token", envToken, "API token for authentication")
	agentID := flag.String("agent-id", envAgentID, "Agent ID (defaults to hostname)")
	interval := flag.Duration("interval", defaultInterval, "Collection interval")
	filterMode := flag.String("filter-mode", defaultFilterMode, "Filter mode: none, basic, aggressive")
	excludePatterns := flag.String("exclude-patterns", envExcludePatterns, "Comma-separated patterns to exclude")
	logLevel := flag.String("log-level", defaultLogLevel, "Log level: debug, info, warn, error")
	showVersion := flag.Bool("version", false, "Print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Println(osqueryagent.Version)
		os.Exit(0)
	}

	// Load config file if specified
	if *configFile != "" {
		cfg, err := loadConfig(*configFile)
		if err != nil {
			log.Fatal().Err(err).Str("file", *configFile).Msg("Failed to load config file")
		}
		// Config file overrides env vars but not explicit flags
		if *serverURL == "" && cfg.Server.URL != "" {
			*serverURL = cfg.Server.URL
		}
		if *apiToken == "" && cfg.Server.APIToken != "" {
			*apiToken = cfg.Server.APIToken
		}
		if *agentID == "" && cfg.Agent.ID != "" {
			*agentID = cfg.Agent.ID
		}
		if *interval == defaultInterval && cfg.Agent.Interval != 0 {
			*interval = cfg.Agent.Interval
		}
		if *filterMode == defaultFilterMode && cfg.Filter.Mode != "" {
			*filterMode = cfg.Filter.Mode
		}
		if *excludePatterns == "" && len(cfg.Filter.ExcludePatterns) > 0 {
			*excludePatterns = strings.Join(cfg.Filter.ExcludePatterns, ",")
		}
		if *logLevel == defaultLogLevel && cfg.Logging.Debug {
			*logLevel = "debug"
		}
	}

	// Validate and set defaults
	if *serverURL == "" {
		*serverURL = "http://localhost:7655"
	}

	if *apiToken == "" {
		fmt.Fprintln(os.Stderr, "error: API token is required (via --api-token, PULSE_TOKEN, or config file)")
		os.Exit(1)
	}

	// Parse log level
	parsedLogLevel, err := parseLogLevel(*logLevel)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// Setup logging
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.SetGlobalLevel(parsedLogLevel)
	if parsedLogLevel == zerolog.DebugLevel {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})
	} else {
		log.Logger = zerolog.New(os.Stdout).Level(parsedLogLevel).With().Timestamp().Logger()
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
		Str("version", osqueryagent.Version).
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

	// Health check
	log.Info().Msg("Performing health check...")
	if err := agent.HealthCheck(); err != nil {
		log.Warn().Err(err).Msg("Health check failed, will retry during operation")
	} else {
		log.Info().Msg("Health check passed")
	}

	// Register agent
	log.Info().Msg("Registering with Pulse server...")
	if err := agent.Register(); err != nil {
		log.Warn().Err(err).Msg("Registration failed, will retry during operation")
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

func parseLogLevel(value string) (zerolog.Level, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return zerolog.InfoLevel, nil
	}

	level, err := zerolog.ParseLevel(normalized)
	if err != nil {
		return zerolog.InfoLevel, fmt.Errorf("invalid log level %q: must be debug, info, warn, or error", value)
	}

	return level, nil
}
