#!/bin/bash
# Run osquery tests
go test -v -tags=test ./internal/api -run Osquery
go test -v ./internal/osqueryagent/...
