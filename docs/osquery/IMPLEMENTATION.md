# osquery Integration - Implementation Documentation

## Overview

The osquery integration enables Pulse to collect and monitor running processes, services, and their associated log files from Proxmox VMs and LXC containers using osquery's powerful SQL-based querying capabilities.

## Architecture

### Components

```
┌─────────────────────────────────────────────────────────────┐
│                     Pulse Agent (VM/LXC)                     │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  osqueryagent Package                                   │ │
│  │  ├── agent.go       - Main agent loop                  │ │
│  │  ├── collector.go   - Data collection via osqueryi     │ │
│  │  │   ├── collectProcesses()                            │ │
│  │  │   ├── collectServices()                             │ │
│  │  │   └── collectOpenFiles()                            │ │
│  │  ├── client.go      - HTTP client to Pulse server      │ │
│  │  └── version.go     - Version tracking                 │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ POST /api/agents/{id}/osquery
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                      Pulse Server                            │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  API Layer (internal/api/)                             │ │
│  │  ├── osquery_agents.go                                 │ │
│  │  │   ├── HandleReport()      - Receive data (POST)     │ │
│  │  │   └── HandleAllReports()  - Retrieve data (GET)     │ │
│  │  └── osquery_store.go                                  │ │
│  │      ├── SaveReport()        - Store in SQLite         │ │
│  │      ├── GetLatestReport()   - Get single agent        │ │
│  │      └── GetAllLatestReports() - Get all agents        │ │
│  └────────────────────────────────────────────────────────┘ │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Storage (SQLite)                                       │ │
│  │  Database: {dataPath}/osquery.db                       │ │
│  │  Table: osquery_reports                                │ │
│  │    - id, agent_id, timestamp                           │ │
│  │    - processes (JSON), services (JSON)                 │ │
│  └────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘
                              │
                              │ GET /api/osquery/reports
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Frontend (React/TS)                       │
│  ┌────────────────────────────────────────────────────────┐ │
│  │  Components (frontend-modern/src/components/Osquery/)  │ │
│  │  ├── OsqueryHosts.tsx          - Main view             │ │
│  │  ├── OsqueryHostSummaryTable   - VM/LXC list           │ │
│  │  ├── OsqueryUnifiedTable       - Process details       │ │
│  │  ├── OsqueryFilter             - Search/filter         │ │
│  │  └── OsqueryStatusBadge        - Status indicators     │ │
│  └────────────────────────────────────────────────────────┘ │
│  Route: /osquery                                            │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow

### Collection Flow

1. **Agent Initialization**
   - Agent starts with `--enable-osquery` flag
   - Checks for `osqueryi` binary availability
   - Initializes HTTP client with Pulse server URL and token

2. **Data Collection** (every interval, default 30s)
   ```go
   // 1. Collect running processes
   processes := collectProcesses()
   // Query: SELECT pid, name, path FROM processes;
   
   // 2. Collect services by OS
   services := collectServices()
   // Linux:   SELECT id, active_state, sub_state FROM systemd_units WHERE id LIKE '%.service';
   // Windows: SELECT name, state, status FROM services;
   // macOS:   SELECT name, state, status FROM launchd WHERE type='service';
   
   // 3. Collect open files per process
   openFiles := collectOpenFiles()
   // Query: SELECT pid, path FROM process_open_files WHERE path != '';
   
   // 4. Filter log files by extension
   logFiles := filterLogFiles(openFiles)
   // Extensions: .log, .txt, .out, .err
   
   // 5. Attach log files to processes
   for each process:
       process.LogFiles = logFiles[process.PID]
   ```

3. **Data Transmission**
   ```json
   POST /api/agents/{agentID}/osquery
   {
     "processes": [
       {
         "pid": "1234",
         "name": "nginx",
         "path": "/usr/sbin/nginx",
         "log_files": ["/var/log/nginx/access.log", "/var/log/nginx/error.log"]
       }
     ],
     "services": [
       {
         "name": "nginx.service",
         "state": "active",
         "status": "running"
       }
     ],
     "timestamp": "2024-01-15T10:30:00Z"
   }
   ```

4. **Server Storage**
   - Validates authentication (Bearer token)
   - Parses JSON payload
   - Stores in SQLite with indexed `agent_id` and `timestamp`
   - Returns success response

### Retrieval Flow

1. **Frontend Request**
   ```typescript
   const reports = await OsqueryAPI.getAllReports();
   // GET /api/osquery/reports
   ```

2. **Server Query**
   ```sql
   SELECT agent_id, timestamp, processes, services 
   FROM osquery_reports r1
   WHERE timestamp = (
     SELECT MAX(timestamp) 
     FROM osquery_reports r2 
     WHERE r2.agent_id = r1.agent_id
   )
   GROUP BY agent_id
   ```

3. **Response Format**
   ```json
   {
     "agent-123": {
       "timestamp": "2024-01-15T10:30:00Z",
       "processes": [...],
       "services": [...]
     },
     "agent-456": {
       "timestamp": "2024-01-15T10:30:15Z",
       "processes": [...],
       "services": [...]
     }
   }
   ```

## Database Schema

### osquery_reports Table

```sql
CREATE TABLE osquery_reports (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  agent_id TEXT NOT NULL,
  timestamp DATETIME NOT NULL,
  processes TEXT,  -- JSON array
  services TEXT,   -- JSON array
  created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_agent_timestamp 
ON osquery_reports(agent_id, timestamp DESC);
```

**Storage Location**: `{dataPath}/osquery.db`

**Retention**: No automatic cleanup (implement if needed)

## API Endpoints

### POST /api/agents/{agentID}/osquery

**Purpose**: Receive osquery data from agents

**Authentication**: Bearer token with `host:report` scope

**Request Body**:
```json
{
  "processes": [
    {
      "pid": "string",
      "name": "string",
      "path": "string",
      "log_files": ["string"]
    }
  ],
  "services": [
    {
      "name": "string",
      "state": "string",
      "status": "string"
    }
  ],
  "timestamp": "RFC3339 string"
}
```

**Response**:
```json
{
  "success": true,
  "message": "Report received"
}
```

### GET /api/agents/{agentID}/osquery

**Purpose**: Retrieve latest report for specific agent

**Authentication**: Bearer token with `monitoring:read` scope

**Response**:
```json
{
  "timestamp": "2024-01-15T10:30:00Z",
  "processes": [...],
  "services": [...]
}
```

### GET /api/osquery/reports

**Purpose**: Retrieve latest reports for all agents

**Authentication**: Bearer token with `monitoring:read` scope

**Response**:
```json
{
  "agent-id-1": {
    "timestamp": "2024-01-15T10:30:00Z",
    "processes": [...],
    "services": [...]
  },
  "agent-id-2": {...}
}
```

## Configuration

### Agent Configuration

**Environment Variables**:
- `PULSE_ENABLE_OSQUERY=true` - Enable osquery collection
- `PULSE_URL` - Pulse server URL
- `PULSE_TOKEN` - API token with `host:report` scope
- `PULSE_INTERVAL` - Collection interval (default: 30s)

**Command Line**:
```bash
pulse-agent \
  --enable-osquery \
  --url https://pulse.example.com \
  --token <api-token> \
  --interval 30s
```

### Server Configuration

No additional configuration required. The osquery store is automatically initialized when the first report is received.

**Data Path**: Configured via `PULSE_DATA_PATH` (default: `/var/lib/pulse`)

## Error Handling

### Agent Errors

1. **osqueryi Not Found**
   - Agent logs warning: "osqueryi binary not found"
   - Collection skipped, agent continues running
   - Check: `which osqueryi` or install osquery

2. **Query Execution Failed**
   - Logs error with query details
   - Returns empty result set
   - Agent continues with next collection cycle

3. **Network Error**
   - Retries on next interval
   - Logs connection failure
   - No data loss (stateless collection)

### Server Errors

1. **Database Initialization Failed**
   - Logs error on startup
   - Returns 503 Service Unavailable
   - Check disk space and permissions

2. **Invalid JSON Payload**
   - Returns 400 Bad Request
   - Logs parse error
   - Agent should retry with valid data

3. **Database Write Failed**
   - Logs error but returns 200 OK
   - Data lost for this cycle
   - Check disk space

## Performance Considerations

### Agent Performance

- **CPU Impact**: Minimal (~1-2% during collection)
- **Memory**: ~10-20MB per agent
- **Network**: ~5-50KB per report (depends on process count)
- **Disk I/O**: Read-only queries to osquery tables

### Server Performance

- **Database Size**: ~1KB per report
- **Query Performance**: Indexed by agent_id and timestamp
- **Concurrent Writes**: SQLite handles multiple agents
- **Memory**: Minimal (streaming JSON parsing)

### Optimization Tips

1. **Increase Collection Interval**
   ```bash
   pulse-agent --enable-osquery --interval 60s
   ```

2. **Limit Process Count** (future enhancement)
   - Filter by CPU/memory usage
   - Exclude system processes

3. **Database Maintenance**
   ```sql
   -- Delete old reports (keep last 7 days)
   DELETE FROM osquery_reports 
   WHERE timestamp < datetime('now', '-7 days');
   
   -- Vacuum to reclaim space
   VACUUM;
   ```

## Security Considerations

### Authentication

- All API endpoints require authentication
- Agent uses Bearer token with `host:report` scope
- Frontend uses session or API token with `monitoring:read` scope

### Data Privacy

- Process names and paths are collected
- No process arguments (may contain sensitive data)
- Log file paths only (not contents)
- No network connections or file contents

### Network Security

- HTTPS recommended for production
- Token transmitted in Authorization header
- No sensitive data in URL parameters

## Testing

### Unit Tests

```bash
# Test agent collection
go test ./internal/osqueryagent/...

# Test API handlers
go test ./internal/api/... -run TestOsquery

# Test database operations
go test ./internal/api/... -run TestOsqueryStore
```

### Integration Tests

```bash
# Start test agent
pulse-agent --enable-osquery --url http://localhost:7655 --token test-token

# Verify data collection
curl -H "Authorization: Bearer admin-token" \
  http://localhost:7655/api/osquery/reports

# Check database
sqlite3 /var/lib/pulse/osquery.db "SELECT COUNT(*) FROM osquery_reports;"
```

### Manual Testing

1. **Install osquery on test VM**
   ```bash
   # Ubuntu/Debian
   wget https://pkg.osquery.io/deb/osquery_5.10.2-1.linux_amd64.deb
   sudo dpkg -i osquery_5.10.2-1.linux_amd64.deb
   
   # Verify installation
   osqueryi --version
   ```

2. **Start agent with osquery enabled**
   ```bash
   pulse-agent --enable-osquery --url http://pulse-server:7655 --token <token>
   ```

3. **Verify data in UI**
   - Navigate to `/osquery` route
   - Check host summary table
   - Verify process list with log files

## Troubleshooting

### Agent Issues

**Problem**: "osqueryi binary not found"
```bash
# Solution: Install osquery
curl -L https://pkg.osquery.io/deb/osquery_5.10.2-1.linux_amd64.deb -o osquery.deb
sudo dpkg -i osquery.deb
```

**Problem**: "Failed to send report: connection refused"
```bash
# Solution: Check Pulse server URL and network connectivity
curl -v http://pulse-server:7655/api/health
```

**Problem**: "Authentication failed"
```bash
# Solution: Verify API token has correct scope
# Token must have 'host:report' scope
```

### Server Issues

**Problem**: "Store not initialized"
```bash
# Solution: Check data directory permissions
ls -la /var/lib/pulse/
sudo chown -R pulse:pulse /var/lib/pulse/
```

**Problem**: "Database locked"
```bash
# Solution: Check for concurrent writes or stale locks
lsof /var/lib/pulse/osquery.db
```

### UI Issues

**Problem**: "No osquery agents reporting"
```bash
# Solution: Verify agents are running and sending data
# Check server logs for incoming reports
journalctl -u pulse -f | grep osquery
```

## Future Enhancements

### Planned Features

1. **Query Customization**
   - Allow custom osquery queries via UI
   - Save query templates
   - Schedule custom queries

2. **Alerting**
   - Alert on specific processes (e.g., crypto miners)
   - Alert on service state changes
   - Alert on high process count

3. **Historical Analysis**
   - Process lifecycle tracking
   - Service uptime statistics
   - Log file growth monitoring

4. **Performance Optimization**
   - Incremental updates (only changed processes)
   - Compression for large reports
   - Batch processing for multiple agents

5. **Advanced Filtering**
   - Filter by process name/path
   - Filter by CPU/memory usage
   - Filter by user/group

### Contributing

To contribute to osquery integration:

1. Review architecture documentation
2. Follow existing code patterns
3. Add tests for new features
4. Update documentation
5. Submit pull request

## References

- [osquery Documentation](https://osquery.io/docs/)
- [osquery Schema](https://osquery.io/schema/)
- [Pulse Agent Documentation](../UNIFIED_AGENT.md)
- [Pulse API Documentation](../API.md)
