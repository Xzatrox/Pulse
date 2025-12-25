package api

import (
	"testing"
	"time"
)

func TestMockOsqueryStore_SaveAndRetrieve(t *testing.T) {
	store := NewMockOsqueryStore()
	defer store.Close()

	processes := []map[string]string{{"pid": "1", "name": "init"}}
	services := []map[string]string{{"name": "sshd", "state": "running"}}
	timestamp := time.Now()

	err := store.SaveReport("agent-1", processes, services, timestamp)
	if err != nil {
		t.Errorf("SaveReport failed: %v", err)
	}

	report, err := store.GetLatestReport("agent-1")
	if err != nil {
		t.Errorf("GetLatestReport failed: %v", err)
	}
	if report == nil {
		t.Error("expected report, got nil")
	}
}

func TestMockOsqueryStore_GetAllLatestReports(t *testing.T) {
	store := NewMockOsqueryStore()
	defer store.Close()

	store.SaveReport("agent-1", []interface{}{}, []interface{}{}, time.Now())
	store.SaveReport("agent-2", []interface{}{}, []interface{}{}, time.Now())

	reports, err := store.GetAllLatestReports()
	if err != nil {
		t.Errorf("GetAllLatestReports failed: %v", err)
	}
	if len(reports) != 2 {
		t.Errorf("expected 2 reports, got %d", len(reports))
	}
}
