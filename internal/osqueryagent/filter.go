package osqueryagent

import "strings"

func (a *Agent) filterProcesses(processes []Process) []Process {
	if len(a.cfg.ExcludePatterns) == 0 {
		return processes
	}

	filtered := make([]Process, 0, len(processes))
	for _, p := range processes {
		if !a.shouldExclude(p.Name) {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

func (a *Agent) filterServices(services []Service) []Service {
	if len(a.cfg.ExcludePatterns) == 0 {
		return services
	}

	filtered := make([]Service, 0, len(services))
	for _, s := range services {
		if !a.shouldExclude(s.Name) {
			filtered = append(filtered, s)
		}
	}
	return filtered
}

func (a *Agent) shouldExclude(name string) bool {
	nameLower := strings.ToLower(name)
	for _, pattern := range a.cfg.ExcludePatterns {
		patternLower := strings.ToLower(pattern)
		
		// Wildcard matching
		if strings.HasPrefix(patternLower, "*") && strings.HasSuffix(patternLower, "*") {
			if strings.Contains(nameLower, strings.Trim(patternLower, "*")) {
				return true
			}
		} else if strings.HasPrefix(patternLower, "*") {
			if strings.HasSuffix(nameLower, strings.TrimPrefix(patternLower, "*")) {
				return true
			}
		} else if strings.HasSuffix(patternLower, "*") {
			if strings.HasPrefix(nameLower, strings.TrimSuffix(patternLower, "*")) {
				return true
			}
		} else if nameLower == patternLower {
			return true
		}
	}
	return false
}
