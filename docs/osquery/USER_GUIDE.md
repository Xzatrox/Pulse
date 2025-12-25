# osquery Integration - User Guide

## What is osquery?

osquery is an open-source tool that allows you to query your operating system like a database. It exposes system information as SQL tables, making it easy to monitor processes, services, network connections, and more.

Pulse integrates osquery to provide real-time visibility into what's running on your Proxmox VMs and LXC containers.

## Features

### What You Can Monitor

- **Running Processes**: See all active processes with PID, name, and path
- **System Services**: Monitor service status (active, inactive, failed)
- **Log Files**: Automatically discover log files opened by each process
- **Real-time Updates**: Data refreshes every 30 seconds
- **Multi-Host View**: Monitor multiple VMs/LXCs from a single dashboard

### Use Cases

1. **Security Monitoring**
   - Detect unauthorized processes
   - Monitor for crypto miners or malware
   - Track service state changes

2. **Troubleshooting**
   - Identify which processes are writing to logs
   - Find services that failed to start
   - Locate application log files quickly

3. **Capacity Planning**
   - Track process count over time
   - Monitor service availability
   - Identify resource-intensive applications

4. **Compliance**
   - Verify required services are running
   - Audit installed software
   - Track process execution history

## Quick Start

### Prerequisites

1. **Pulse Server** running v5.0 or later
2. **Pulse Agent** installed on target VMs/LXCs
3. **osquery** installed on target systems

### Step 1: Install osquery

#### Ubuntu/Debian
```bash
# Download osquery package
wget https://pkg.osquery.io/deb/osquery_5.10.2-1.linux_amd64.deb

# Install
sudo dpkg -i osquery_5.10.2-1.linux_amd64.deb

# Verify installation
osqueryi --version
```

#### CentOS/RHEL
```bash
# Download osquery package
wget https://pkg.osquery.io/rpm/osquery-5.10.2-1.linux.x86_64.rpm

# Install
sudo rpm -ivh osquery-5.10.2-1.linux.x86_64.rpm

# Verify installation
osqueryi --version
```

#### Alpine Linux (LXC containers)
```bash
# Install from Alpine repositories
apk add osquery

# Verify installation
osqueryi --version
```

#### Windows
```powershell
# Download installer from https://osquery.io/downloads
# Run installer and follow prompts

# Verify installation
osqueryi.exe --version
```

### Step 2: Enable osquery in Pulse Agent

#### Option A: Environment Variable
```bash
# Edit agent configuration
sudo nano /etc/pulse/agent.env

# Add this line
PULSE_ENABLE_OSQUERY=true

# Restart agent
sudo systemctl restart pulse-agent
```

#### Option B: Command Line
```bash
# Start agent with osquery enabled
pulse-agent \
  --enable-osquery \
  --url https://pulse.example.com \
  --token <your-api-token>
```

#### Option C: Systemd Service
```bash
# Edit systemd service file
sudo systemctl edit pulse-agent

# Add override
[Service]
Environment="PULSE_ENABLE_OSQUERY=true"

# Reload and restart
sudo systemctl daemon-reload
sudo systemctl restart pulse-agent
```

### Step 3: Verify Data Collection

1. **Check Agent Logs**
   ```bash
   # View agent logs
   sudo journalctl -u pulse-agent -f
   
   # Look for osquery messages
   # Should see: "Collected X processes, Y services"
   ```

2. **Check Pulse UI**
   - Navigate to `https://pulse.example.com/osquery`
   - You should see your VM/LXC in the host summary table
   - Click to view detailed process list

## Using the osquery Dashboard

### Host Summary Table

The top table shows all VMs/LXCs reporting osquery data:

| Column | Description |
|--------|-------------|
| Agent ID | Unique identifier for the VM/LXC |
| Processes | Number of running processes |
| Services | Number of monitored services |
| Last Update | Timestamp of last data collection |

### Process Table

The main table shows detailed process information:

| Column | Description |
|--------|-------------|
| PID | Process ID |
| Name | Process name (e.g., nginx, postgres) |
| Path | Full path to executable |
| Log Files | List of log files opened by this process |
| Agent | Which VM/LXC this process is running on |

### Search and Filter

**Search Box**: Filter processes by name
```
Example: Type "nginx" to show only nginx processes
```

**Status Filter**: Filter by service status
- All Status: Show everything
- Running: Show only active services
- Stopped: Show only inactive services

### Understanding the Data

#### Process Information

- **PID**: Unique process identifier (changes on restart)
- **Name**: Short process name (may be truncated)
- **Path**: Full executable path (useful for identifying versions)
- **Log Files**: Files currently open for writing (filtered by .log, .txt, .out, .err extensions)

#### Service Information

- **Name**: Service name (e.g., nginx.service)
- **State**: Service state (active, inactive, failed)
- **Status**: Detailed status (running, stopped, dead)

#### Status Badges

- 🟢 **Active/Running**: Service is running normally
- ⚪ **Inactive/Stopped**: Service is not running
- 🔴 **Failed**: Service failed to start or crashed

## Common Tasks

### Find Log Files for an Application

1. Navigate to `/osquery` in Pulse UI
2. Use search box to filter by application name
3. Check "Log Files" column for discovered logs
4. Click log path to copy (future feature)

**Example**: Finding nginx logs
```
1. Type "nginx" in search box
2. Look at "Log Files" column
3. See: /var/log/nginx/access.log, /var/log/nginx/error.log
```

### Monitor Service Status

1. Navigate to `/osquery` in Pulse UI
2. Look at "Services" section (if implemented)
3. Check status badges for service health
4. Filter by "Stopped" to find failed services

### Track Process Count

1. Check "Processes" column in host summary
2. Compare across multiple VMs/LXCs
3. Identify systems with unusually high process counts
4. Investigate high-count systems for issues

### Identify Unknown Processes

1. Review process list for unfamiliar names
2. Check process path for suspicious locations
3. Cross-reference with known applications
4. Investigate processes in unusual directories (e.g., /tmp)

## Advanced Configuration

### Adjust Collection Interval

**Default**: 30 seconds

**Change Interval**:
```bash
# Collect every 60 seconds
pulse-agent --enable-osquery --interval 60s

# Or via environment variable
PULSE_INTERVAL=60s
```

**Considerations**:
- Lower interval = more frequent updates, higher CPU usage
- Higher interval = less frequent updates, lower CPU usage
- Recommended: 30-60 seconds for most use cases

### Custom osquery Queries (Future)

Currently, Pulse uses predefined queries. Custom queries will be supported in a future release.

**Planned Queries**:
- Network connections
- Installed packages
- User accounts
- File integrity monitoring
- Registry keys (Windows)

### Integration with Alerts (Future)

Pulse will support alerting based on osquery data:

**Planned Alerts**:
- Process detected (e.g., crypto miner)
- Service state changed (e.g., nginx stopped)
- High process count (e.g., >500 processes)
- Suspicious process path (e.g., /tmp/malware)

## Troubleshooting

### No Data Appearing in UI

**Symptom**: `/osquery` page shows "No osquery agents reporting"

**Solutions**:

1. **Verify osquery is installed**
   ```bash
   which osqueryi
   # Should output: /usr/bin/osqueryi
   ```

2. **Check agent is running with osquery enabled**
   ```bash
   ps aux | grep pulse-agent
   # Should see --enable-osquery flag
   ```

3. **Check agent logs for errors**
   ```bash
   sudo journalctl -u pulse-agent -n 50
   # Look for "osqueryi binary not found" or other errors
   ```

4. **Verify network connectivity**
   ```bash
   curl -v https://pulse.example.com/api/health
   # Should return 200 OK
   ```

5. **Check API token permissions**
   - Token must have `host:report` scope
   - Verify in Pulse UI: Settings → API Tokens

### Data Not Updating

**Symptom**: Data is stale (Last Update timestamp is old)

**Solutions**:

1. **Check agent is still running**
   ```bash
   sudo systemctl status pulse-agent
   ```

2. **Restart agent**
   ```bash
   sudo systemctl restart pulse-agent
   ```

3. **Check for network issues**
   ```bash
   # Test connectivity to Pulse server
   curl -H "Authorization: Bearer <token>" \
     https://pulse.example.com/api/agents/<agent-id>/osquery
   ```

### Missing Log Files

**Symptom**: Process shows "None" in Log Files column

**Possible Reasons**:

1. **Process doesn't write to log files**
   - Some processes log to syslog instead
   - Some processes don't log at all

2. **Log files not open at collection time**
   - Process may open/close log files as needed
   - Try refreshing after a few minutes

3. **Log files have non-standard extensions**
   - Currently filters: .log, .txt, .out, .err
   - Other extensions not detected

4. **Permissions issue**
   - osquery may not have permission to read process file descriptors
   - Run agent as root or with appropriate capabilities

### High CPU Usage

**Symptom**: Agent using excessive CPU

**Solutions**:

1. **Increase collection interval**
   ```bash
   # Change from 30s to 60s
   PULSE_INTERVAL=60s
   ```

2. **Check osquery performance**
   ```bash
   # Test query performance
   time osqueryi "SELECT * FROM processes;"
   ```

3. **Reduce process count** (if very high)
   - Investigate why so many processes are running
   - Consider filtering in future release

### Database Growing Too Large

**Symptom**: `/var/lib/pulse/osquery.db` is very large

**Solutions**:

1. **Check database size**
   ```bash
   du -h /var/lib/pulse/osquery.db
   ```

2. **Clean old data** (manual for now)
   ```bash
   sqlite3 /var/lib/pulse/osquery.db
   
   # Delete reports older than 7 days
   DELETE FROM osquery_reports 
   WHERE timestamp < datetime('now', '-7 days');
   
   # Reclaim space
   VACUUM;
   
   # Exit
   .quit
   ```

3. **Implement retention policy** (future feature)
   - Automatic cleanup of old data
   - Configurable retention period

## Best Practices

### Security

1. **Use HTTPS**: Always use HTTPS for Pulse server in production
2. **Rotate Tokens**: Regularly rotate API tokens
3. **Limit Scope**: Use tokens with minimal required scopes
4. **Monitor Access**: Review agent access logs regularly

### Performance

1. **Start with Default Interval**: Use 30s interval initially
2. **Monitor Resource Usage**: Check CPU/memory impact
3. **Adjust as Needed**: Increase interval if resources are constrained
4. **Scale Gradually**: Enable on a few VMs first, then expand

### Maintenance

1. **Update osquery**: Keep osquery updated for security patches
2. **Monitor Database Size**: Check disk usage periodically
3. **Clean Old Data**: Remove old reports to save space
4. **Review Logs**: Check agent logs for errors

### Monitoring

1. **Set Up Alerts**: Configure alerts for critical services (future)
2. **Regular Reviews**: Review process lists weekly
3. **Baseline Normal**: Understand what's normal for your systems
4. **Investigate Anomalies**: Check unusual processes or services

## FAQ

### Q: Does osquery impact system performance?

**A**: Minimal impact. osquery queries are read-only and optimized. Typical CPU usage is 1-2% during collection.

### Q: Can I use osquery without Pulse Agent?

**A**: Yes, osquery is standalone. Pulse Agent just collects and sends data to Pulse server.

### Q: What data is collected?

**A**: Process names, PIDs, paths, service status, and log file paths. No process arguments, file contents, or network data.

### Q: How long is data retained?

**A**: Currently indefinite. Manual cleanup required. Automatic retention policy coming in future release.

### Q: Can I query historical data?

**A**: Not yet. Currently shows latest snapshot only. Historical analysis planned for future release.

### Q: Does this work on Windows?

**A**: Yes, osquery supports Windows. Pulse Agent with osquery support works on Windows, Linux, and macOS.

### Q: Can I customize the queries?

**A**: Not yet. Custom queries planned for future release. Currently uses predefined queries for processes and services.

### Q: How do I disable osquery collection?

**A**: Remove `--enable-osquery` flag or set `PULSE_ENABLE_OSQUERY=false`, then restart agent.

### Q: Is osquery required for Pulse Agent?

**A**: No, osquery is optional. Agent works without it. Only enable if you need process/service monitoring.

### Q: Can I monitor Docker containers?

**A**: Yes, if osquery is installed in the container. Works same as VMs/LXCs.

### Q: What's the difference between osquery and Pulse monitoring?

**A**: Pulse monitors infrastructure (CPU, memory, disk). osquery monitors processes and services. They complement each other.

## Getting Help

### Documentation

- [Implementation Documentation](IMPLEMENTATION.md) - Technical details
- [Pulse Documentation](../README.md) - General Pulse docs
- [osquery Documentation](https://osquery.io/docs/) - Official osquery docs

### Support

- **GitHub Issues**: [Report bugs or request features](https://github.com/rcourtman/Pulse/issues)
- **Discussions**: [Ask questions](https://github.com/rcourtman/Pulse/discussions)
- **Slack**: [Join #osctrl channel](https://osquery.slack.com)

### Contributing

Contributions welcome! See [CONTRIBUTING.md](../../CONTRIBUTING.md) for guidelines.

## Changelog

### v5.0.0 (Initial Release)
- ✅ Process monitoring
- ✅ Service monitoring
- ✅ Log file discovery
- ✅ Multi-host dashboard
- ✅ Search and filter
- ✅ Real-time updates

### Planned Features
- 🔄 Custom queries
- 🔄 Historical analysis
- 🔄 Alerting integration
- 🔄 Performance metrics
- 🔄 Export to CSV
- 🔄 Process lifecycle tracking

## License

osquery integration is part of Pulse and licensed under MIT License. See [LICENSE](../../LICENSE) for details.

osquery itself is licensed under Apache 2.0 or GPL 2.0. See [osquery license](https://github.com/osquery/osquery/blob/master/LICENSE) for details.
