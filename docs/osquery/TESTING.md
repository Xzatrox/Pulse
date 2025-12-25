# Osquery Testing Guide

## Run Tests

```bash
# Test API handlers
go test -v ./internal/api -run Osquery

# Test agent module
go test -v ./internal/osqueryagent/...

# All tests
go test -v ./internal/api/... ./internal/osqueryagent/...
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
