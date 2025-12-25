# Osquery Testing Guide

## Run Tests

```bash
# Test API handlers
go test ./internal/api/osquery_agents_test.go ./internal/api/osquery_store_test.go ./internal/api/osquery_agents.go ./internal/api/osquery_store.go -v

# Test agent module
go test ./internal/osqueryagent/... -v

# All osquery tests
go test ./internal/api/osquery*.go ./internal/osqueryagent/... -v
```

## Manual Testing

1. Start agent with osquery enabled:
```bash
./pulse-agent --enable-osquery --pulse-url http://localhost:7655 --api-token <token>
```

2. Verify reports endpoint:
```bash
curl -H "Authorization: Bearer <token>" http://localhost:7655/api/osquery/reports
```

3. Test frontend at: http://localhost:7655/osquery
