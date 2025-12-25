package api

import (
	"sync"
	"time"
)

type MockOsqueryStore struct {
	mu      sync.RWMutex
	reports map[string]map[string]interface{}
}

func NewMockOsqueryStore() *MockOsqueryStore {
	return &MockOsqueryStore{
		reports: make(map[string]map[string]interface{}),
	}
}

func (m *MockOsqueryStore) SaveReport(agentID string, processes, services interface{}, timestamp time.Time) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.reports[agentID] = map[string]interface{}{
		"timestamp": timestamp,
		"processes": processes,
		"services":  services,
	}
	return nil
}

func (m *MockOsqueryStore) GetLatestReport(agentID string) (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	report, ok := m.reports[agentID]
	if !ok {
		return nil, nil
	}
	return report, nil
}

func (m *MockOsqueryStore) GetAllLatestReports() (map[string]interface{}, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	result := make(map[string]interface{})
	for agentID, report := range m.reports {
		result[agentID] = report
	}
	return result, nil
}

func (m *MockOsqueryStore) Close() error {
	return nil
}
