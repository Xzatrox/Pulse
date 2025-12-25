package osqueryagent

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Report struct {
	Processes []Process `json:"processes"`
	Services  []Service `json:"services"`
	Timestamp time.Time `json:"timestamp"`
}

func (a *Agent) sendReport(processes []Process, services []Service) error {
	report := Report{
		Processes: processes,
		Services:  services,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(report)
	if err != nil {
		return err
	}

	url := fmt.Sprintf("%s/api/agents/%s/osquery", a.cfg.PulseURL, a.cfg.AgentID)
	req, err := http.NewRequest("POST", url, bytes.NewReader(data))
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+a.cfg.APIToken)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{
		Timeout: 30 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: a.cfg.InsecureSkipVerify,
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("server returned %d", resp.StatusCode)
	}

	return nil
}
