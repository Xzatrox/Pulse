package osqueryagent

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type RegistrationRequest struct {
	AgentID  string `json:"agent_id"`
	Hostname string `json:"hostname"`
	Version  string `json:"version"`
}

type RegistrationResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func (a *Agent) Register() error {
	hostname, _ := getHostname()
	
	req := RegistrationRequest{
		AgentID:  a.cfg.AgentID,
		Hostname: hostname,
		Version:  Version,
	}
	
	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal registration: %w", err)
	}
	
	client := &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: a.cfg.InsecureSkipVerify,
			},
		},
	}
	
	url := fmt.Sprintf("%s/api/osquery/register", a.cfg.PulseURL)
	httpReq, err := http.NewRequest("POST", url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	
	httpReq.Header.Set("Content-Type", "application/json")
	if a.cfg.APIToken != "" {
		httpReq.Header.Set("Authorization", "Bearer "+a.cfg.APIToken)
	}
	
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send registration: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("registration failed with status %d", resp.StatusCode)
	}
	
	var regResp RegistrationResponse
	if err := json.NewDecoder(resp.Body).Decode(&regResp); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	
	if !regResp.Success {
		return fmt.Errorf("registration rejected: %s", regResp.Message)
	}
	
	a.cfg.Logger.Info().Str("agent_id", a.cfg.AgentID).Msg("Successfully registered with Pulse server")
	return nil
}

func (a *Agent) HealthCheck() error {
	client := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: a.cfg.InsecureSkipVerify,
			},
		},
	}
	
	url := fmt.Sprintf("%s/api/health", a.cfg.PulseURL)
	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}
	
	return nil
}

func getHostname() (string, error) {
	return "", nil
}
