# osquery Integration Documentation

Complete documentation for Pulse's osquery integration feature.

## Documentation Index

### 📘 [User Guide](USER_GUIDE.md)
**For end users and administrators**

Learn how to:
- Install and configure osquery
- Enable osquery in Pulse Agent
- Use the osquery dashboard
- Monitor processes and services
- Troubleshoot common issues

**Start here if you're**: Setting up osquery monitoring for the first time

---

### 🔧 [Implementation Documentation](IMPLEMENTATION.md)
**For developers and contributors**

Technical details:
- Architecture and data flow
- API endpoints and schemas
- Database structure
- Performance considerations
- Testing and troubleshooting

**Start here if you're**: Contributing code or need technical details

---

## Quick Links

### Getting Started
1. [Prerequisites](USER_GUIDE.md#prerequisites)
2. [Install osquery](USER_GUIDE.md#step-1-install-osquery)
3. [Enable in agent](USER_GUIDE.md#step-2-enable-osquery-in-pulse-agent)
4. [View dashboard](USER_GUIDE.md#using-the-osquery-dashboard)

### Common Tasks
- [Find log files](USER_GUIDE.md#find-log-files-for-an-application)
- [Monitor services](USER_GUIDE.md#monitor-service-status)
- [Track processes](USER_GUIDE.md#track-process-count)
- [Troubleshoot issues](USER_GUIDE.md#troubleshooting)

### Technical Reference
- [Architecture](IMPLEMENTATION.md#architecture)
- [API Endpoints](IMPLEMENTATION.md#api-endpoints)
- [Database Schema](IMPLEMENTATION.md#database-schema)
- [Data Flow](IMPLEMENTATION.md#data-flow)

## What is osquery?

osquery is an open-source tool that exposes operating system information as SQL tables. Pulse integrates osquery to provide:

- **Process Monitoring**: See what's running on your VMs/LXCs
- **Service Status**: Monitor systemd/Windows services
- **Log Discovery**: Automatically find log files
- **Real-time Updates**: Data refreshes every 30 seconds

## Features

### Current (v5.0)
- ✅ Process monitoring (PID, name, path)
- ✅ Service monitoring (state, status)
- ✅ Log file discovery (by extension)
- ✅ Multi-host dashboard
- ✅ Search and filter
- ✅ Real-time updates

### Planned (Future)
- 🔄 Custom osquery queries
- 🔄 Historical analysis
- 🔄 Alert integration
- 🔄 Performance metrics
- 🔄 Process lifecycle tracking
- 🔄 Network connection monitoring

## Architecture Overview

```
┌─────────────┐
│   VM/LXC    │
│  ┌────────┐ │
│  │osquery │ │
│  └────────┘ │
│  ┌────────┐ │
│  │ Agent  │ │──POST──┐
│  └────────┘ │        │
└─────────────┘        │
                       ▼
┌─────────────┐   ┌─────────────┐
│   VM/LXC    │   │    Pulse    │
│  ┌────────┐ │   │   Server    │
│  │osquery │ │   │  ┌────────┐ │
│  └────────┘ │   │  │ SQLite │ │
│  ┌────────┐ │   │  └────────┘ │
│  │ Agent  │ │──POST─────────┐ │
│  └────────┘ │   │            │ │
└─────────────┘   └────────────┼─┘
                               │
                               ▼
                          ┌─────────┐
                          │   UI    │
                          │/osquery │
                          └─────────┘
```

## Installation

### Quick Start

```bash
# 1. Install osquery on target system
wget https://pkg.osquery.io/deb/osquery_5.10.2-1.linux_amd64.deb
sudo dpkg -i osquery_5.10.2-1.linux_amd64.deb

# 2. Enable in Pulse Agent
pulse-agent --enable-osquery --url https://pulse.example.com --token <token>

# 3. View in UI
# Navigate to https://pulse.example.com/osquery
```

See [User Guide](USER_GUIDE.md#quick-start) for detailed instructions.

## Usage Examples

### Monitor nginx Processes
```
1. Navigate to /osquery in Pulse UI
2. Type "nginx" in search box
3. View running nginx processes and their log files
```

### Find Failed Services
```
1. Navigate to /osquery in Pulse UI
2. Select "Stopped" in status filter
3. Review services that are not running
```

### Track Process Count
```
1. Check "Processes" column in host summary
2. Compare across VMs/LXCs
3. Investigate systems with high counts
```

## API Reference

### Endpoints

```
POST /api/agents/{agentID}/osquery
  - Submit osquery data from agent
  - Auth: Bearer token (host:report scope)

GET /api/agents/{agentID}/osquery
  - Get latest report for agent
  - Auth: Bearer token (monitoring:read scope)

GET /api/osquery/reports
  - Get latest reports for all agents
  - Auth: Bearer token (monitoring:read scope)
```

See [Implementation Documentation](IMPLEMENTATION.md#api-endpoints) for details.

## Configuration

### Agent Configuration

```bash
# Environment variable
export PULSE_ENABLE_OSQUERY=true

# Command line
pulse-agent --enable-osquery

# Systemd service
[Service]
Environment="PULSE_ENABLE_OSQUERY=true"
```

### Server Configuration

No configuration required. Automatically initialized on first report.

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| No data in UI | Verify osquery installed, agent running |
| Data not updating | Restart agent, check network |
| Missing log files | Check file extensions, permissions |
| High CPU usage | Increase collection interval |

See [User Guide - Troubleshooting](USER_GUIDE.md#troubleshooting) for detailed solutions.

## Performance

### Resource Usage

| Component | CPU | Memory | Disk | Network |
|-----------|-----|--------|------|---------|
| Agent | 1-2% | 10-20MB | - | 5-50KB/report |
| Server | <1% | Minimal | ~1KB/report | - |

### Optimization

- Default interval: 30s (adjustable)
- Database indexed by agent_id and timestamp
- Streaming JSON parsing (low memory)
- No data retention limit (manual cleanup)

## Security

### Data Collected

- ✅ Process names and paths
- ✅ Service status
- ✅ Log file paths
- ❌ Process arguments (may contain secrets)
- ❌ File contents
- ❌ Network connections

### Authentication

- All endpoints require authentication
- Agent uses Bearer token with `host:report` scope
- UI uses session or token with `monitoring:read` scope
- HTTPS recommended for production

## Contributing

Contributions welcome! Areas for improvement:

- Custom query support
- Historical analysis
- Alert integration
- Performance optimization
- Additional osquery tables

See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## Support

- **Documentation**: This directory
- **Issues**: [GitHub Issues](https://github.com/rcourtman/Pulse/issues)
- **Discussions**: [GitHub Discussions](https://github.com/rcourtman/Pulse/discussions)
- **Slack**: [#osctrl channel](https://osquery.slack.com)

## License

MIT License - See [LICENSE](../../LICENSE)

osquery is licensed under Apache 2.0 or GPL 2.0 - See [osquery license](https://github.com/osquery/osquery/blob/master/LICENSE)

## Related Documentation

- [Pulse Agent Documentation](../UNIFIED_AGENT.md)
- [Pulse API Documentation](../API.md)
- [osquery Official Documentation](https://osquery.io/docs/)
- [osquery Schema Reference](https://osquery.io/schema/)
