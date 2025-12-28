package osqueryagent

import (
	"encoding/json"
	"os/exec"
	"runtime"
	"strings"
)

type Process struct {
	PID         string   `json:"pid"`
	Name        string   `json:"name"`
	Path        string   `json:"path"`
	LogFiles    []string `json:"log_files,omitempty"`
	MemoryBytes string   `json:"memory_bytes,omitempty"`
}

type Service struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Status string `json:"status"`
}

type OpenFile struct {
	PID  string `json:"pid"`
	Path string `json:"path"`
}

var logExtensions = []string{".log", ".txt", ".out", ".err"}

func (a *Agent) collectProcesses() ([]Process, error) {
	cmd := exec.Command("osqueryi", "--json", "SELECT pid, name, path, resident_size FROM processes;")
	output, err := cmd.Output()
	if err != nil {
		a.cfg.Logger.Warn().Err(err).Msg("Failed to collect process memory, retrying without resident_size")
		cmd = exec.Command("osqueryi", "--json", "SELECT pid, name, path FROM processes;")
		output, err = cmd.Output()
		if err != nil {
			return nil, err
		}
	}
	
	var rawProcesses []struct {
		PID          string `json:"pid"`
		Name         string `json:"name"`
		Path         string `json:"path"`
		ResidentSize string `json:"resident_size"`
	}
	if err := json.Unmarshal(output, &rawProcesses); err != nil {
		return nil, err
	}
	
	processes := make([]Process, len(rawProcesses))
	for i, rp := range rawProcesses {
		processes[i] = Process{
			PID:         rp.PID,
			Name:        rp.Name,
			Path:        rp.Path,
			MemoryBytes: rp.ResidentSize,
		}
	}
	
	openFiles, _ := a.collectOpenFiles()
	for i := range processes {
		processes[i].LogFiles = filterLogFiles(openFiles[processes[i].PID])
	}
	
	return processes, nil
}

func (a *Agent) collectOpenFiles() (map[string][]string, error) {
	var query string
	switch runtime.GOOS {
	case "linux", "darwin":
		query = "SELECT pid, path FROM process_open_files WHERE path != '';"
	case "windows":
		query = "SELECT pid, path FROM process_open_files WHERE path != '';"
	default:
		return nil, nil
	}
	
	cmd := exec.Command("osqueryi", "--json", query)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var files []OpenFile
	json.Unmarshal(output, &files)
	
	filesByPID := make(map[string][]string)
	for _, f := range files {
		filesByPID[f.PID] = append(filesByPID[f.PID], f.Path)
	}
	
	return filesByPID, nil
}

func filterLogFiles(files []string) []string {
	var logs []string
	for _, file := range files {
		for _, ext := range logExtensions {
			if strings.HasSuffix(strings.ToLower(file), ext) {
				logs = append(logs, file)
				break
			}
		}
	}
	return logs
}

func (a *Agent) collectServices() ([]Service, error) {
	var query string
	switch runtime.GOOS {
	case "linux":
		query = "SELECT id as name, active_state as state, sub_state as status FROM systemd_units WHERE id LIKE '%.service';"
	case "windows":
		query = "SELECT name, state, status FROM services;"
	case "darwin":
		query = "SELECT name, state, status FROM launchd WHERE type='service';"
	default:
		return nil, nil
	}
	
	cmd := exec.Command("osqueryi", "--json", query)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	
	var services []Service
	json.Unmarshal(output, &services)
	return services, nil
}
