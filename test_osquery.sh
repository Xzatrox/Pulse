#!/bin/bash
# Run osquery tests
go test ./internal/api/osquery_agents_test.go ./internal/api/osquery_store_test.go ./internal/api/osquery_agents.go ./internal/api/osquery_store.go -v
go test ./internal/osqueryagent/... -v
