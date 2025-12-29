package api

import (
	"database/sql"
	"encoding/json"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/rs/zerolog/log"
)

type OsqueryStore struct {
	db *sql.DB
}

func NewOsqueryStore(dataPath string) (*OsqueryStore, error) {
	dbPath := filepath.Join(dataPath, "osquery.db")
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	store := &OsqueryStore{db: db}
	if err := store.init(); err != nil {
		db.Close()
		return nil, err
	}

	return store, nil
}

func (s *OsqueryStore) init() error {
	schema := `
	CREATE TABLE IF NOT EXISTS osquery_reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		agent_id TEXT NOT NULL,
		timestamp DATETIME NOT NULL,
		processes TEXT,
		services TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);
	CREATE INDEX IF NOT EXISTS idx_agent_timestamp ON osquery_reports(agent_id, timestamp DESC);
	`
	_, err := s.db.Exec(schema)
	return err
}

func (s *OsqueryStore) SaveReport(agentID string, processes, services interface{}, timestamp time.Time) error {
	procJSON, _ := json.Marshal(processes)
	svcJSON, _ := json.Marshal(services)

	_, err := s.db.Exec(
		"INSERT INTO osquery_reports (agent_id, timestamp, processes, services) VALUES (?, ?, ?, ?)",
		agentID, timestamp, procJSON, svcJSON,
	)
	return err
}

func (s *OsqueryStore) GetLatestReport(agentID string) (map[string]interface{}, error) {
	var procJSON, svcJSON string
	var timestamp time.Time

	err := s.db.QueryRow(
		"SELECT timestamp, processes, services FROM osquery_reports WHERE agent_id = ? ORDER BY timestamp DESC LIMIT 1",
		agentID,
	).Scan(&timestamp, &procJSON, &svcJSON)

	if err != nil {
		return nil, err
	}

	var processes, services interface{}
	json.Unmarshal([]byte(procJSON), &processes)
	json.Unmarshal([]byte(svcJSON), &services)

	return map[string]interface{}{
		"timestamp": timestamp,
		"processes": processes,
		"services":  services,
	}, nil
}

func (s *OsqueryStore) GetAllLatestReports() (map[string]interface{}, error) {
	rows, err := s.db.Query(`
		SELECT agent_id, timestamp, processes, services 
		FROM osquery_reports r1
		WHERE timestamp = (SELECT MAX(timestamp) FROM osquery_reports r2 WHERE r2.agent_id = r1.agent_id)
		GROUP BY agent_id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	reports := make(map[string]interface{})
	for rows.Next() {
		var agentID, procJSON, svcJSON string
		var timestamp time.Time
		if err := rows.Scan(&agentID, &timestamp, &procJSON, &svcJSON); err != nil {
			log.Warn().Err(err).Msg("Failed to scan osquery report row")
			continue
		}

		var processes, services interface{}
		json.Unmarshal([]byte(procJSON), &processes)
		json.Unmarshal([]byte(svcJSON), &services)

		reports[agentID] = map[string]interface{}{
			"timestamp": timestamp,
			"processes": processes,
			"services":  services,
		}
	}

	return reports, nil
}

func (s *OsqueryStore) CleanupOldReports(retentionDays int) (int64, error) {
	cutoffTime := time.Now().AddDate(0, 0, -retentionDays)
	result, err := s.db.Exec(
		"DELETE FROM osquery_reports WHERE created_at < ?",
		cutoffTime,
	)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (s *OsqueryStore) Close() error {
	return s.db.Close()
}
