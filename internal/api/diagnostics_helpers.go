package api

import (
	"encoding/json"
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rcourtman/pulse-go-rewrite/internal/alerts"
	"github.com/rcourtman/pulse-go-rewrite/internal/config"
	"github.com/rcourtman/pulse-go-rewrite/internal/models"
	"golang.org/x/crypto/ssh"
)

type HostProxySummary struct {
	Requested              bool   `json:"requested"`
	Installed              bool   `json:"installed"`
	HostSocketPresent      bool   `json:"hostSocketPresent"`
	ContainerSocketPresent *bool  `json:"containerSocketPresent,omitempty"`
	LastUpdated            string `json:"lastUpdated,omitempty"`
	CTID                   string `json:"ctid,omitempty"`
}

func loadHostProxySummary() (*HostProxySummary, error) {
	const summaryPath = "/etc/pulse/install_summary.json"
	data, err := os.ReadFile(summaryPath)
	if err != nil {
		return nil, err
	}
	var raw struct {
		GeneratedAt string `json:"generatedAt"`
		CTID        string `json:"ctid"`
		Proxy       struct {
			Requested              bool  `json:"requested"`
			Installed              bool  `json:"installed"`
			HostSocketPresent      bool  `json:"hostSocketPresent"`
			ContainerSocketPresent *bool `json:"containerSocketPresent"`
		} `json:"proxy"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	summary := &HostProxySummary{
		Requested:         raw.Proxy.Requested,
		Installed:         raw.Proxy.Installed,
		HostSocketPresent: raw.Proxy.HostSocketPresent,
		LastUpdated:       strings.TrimSpace(raw.GeneratedAt),
		CTID:              strings.TrimSpace(raw.CTID),
	}
	if raw.Proxy.ContainerSocketPresent != nil {
		value := *raw.Proxy.ContainerSocketPresent
		summary.ContainerSocketPresent = &value
	}
	return summary, nil
}

func preferredDockerHostName(host models.DockerHost) string {
	if name := strings.TrimSpace(host.DisplayName); name != "" {
		return name
	}
	if name := strings.TrimSpace(host.Hostname); name != "" {
		return name
	}
	if name := strings.TrimSpace(host.AgentID); name != "" {
		return name
	}
	return host.ID
}

func normalizeVersionLabel(raw string) string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return ""
	}
	if strings.HasPrefix(value, "v") {
		return value
	}
	first := value[0]
	if first < '0' || first > '9' {
		return value
	}
	return "v" + value
}

func normalizeHostForComparison(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	trimmed = strings.TrimPrefix(trimmed, "https://")
	trimmed = strings.TrimPrefix(trimmed, "http://")
	if idx := strings.IndexByte(trimmed, '/'); idx != -1 {
		trimmed = trimmed[:idx]
	}
	if idx := strings.IndexByte(trimmed, ':'); idx != -1 {
		trimmed = trimmed[:idx]
	}
	return strings.ToLower(strings.TrimSpace(trimmed))
}

func copyStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return append([]string(nil), values...)
}

func isFallbackMemorySource(source string) bool {
	switch strings.ToLower(source) {
	case "", "unknown", "nodes-endpoint", "node-status-used", "previous-snapshot":
		return true
	default:
		return false
	}
}

func contains(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}

func containsFold(slice []string, candidate string) bool {
	target := strings.ToLower(strings.TrimSpace(candidate))
	if target == "" {
		return false
	}

	for _, s := range slice {
		if strings.ToLower(strings.TrimSpace(s)) == target {
			return true
		}
	}
	return false
}

func interfaceToStringSlice(value interface{}) []string {
	switch v := value.(type) {
	case []string:
		out := make([]string, len(v))
		copy(out, v)
		return out
	case []interface{}:
		result := make([]string, 0, len(v))
		for _, item := range v {
			if str, ok := item.(string); ok {
				result = append(result, str)
			}
		}
		return result
	default:
		return nil
	}
}

func formatTimeMaybe(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func matchInstanceNameByHost(cfg *config.Config, host string) string {
	if cfg == nil {
		return ""
	}
	needle := normalizeHostForComparison(host)
	if needle == "" {
		return ""
	}
	for _, inst := range cfg.PVEInstances {
		candidate := normalizeHostForComparison(inst.Host)
		if candidate != "" && strings.EqualFold(candidate, needle) {
			return strings.TrimSpace(inst.Name)
		}
	}
	return ""
}

func hasLegacyThresholds(th alerts.ThresholdConfig) bool {
	return th.CPULegacy != nil ||
		th.MemoryLegacy != nil ||
		th.DiskLegacy != nil ||
		th.DiskReadLegacy != nil ||
		th.DiskWriteLegacy != nil ||
		th.NetworkInLegacy != nil ||
		th.NetworkOutLegacy != nil
}

func fingerprintPublicKey(pub string) (string, error) {
	pub = strings.TrimSpace(pub)
	if pub == "" {
		return "", errors.New("empty public key")
	}
	key, _, _, _, err := ssh.ParseAuthorizedKey([]byte(pub))
	if err != nil {
		return "", err
	}
	return ssh.FingerprintSHA256(key), nil
}

func countLegacySSHKeys(dir string) (int, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, nil
		}
		return 0, err
	}
	count := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if strings.HasPrefix(name, "id_") {
			count++
		}
	}
	return count, nil
}

func resolveUserName(uid uint32) string {
	return "uid:" + strconv.FormatUint(uint64(uid), 10)
}

func resolveGroupName(gid uint32) string {
	return "gid:" + strconv.FormatUint(uint64(gid), 10)
}
