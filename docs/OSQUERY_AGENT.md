# Pulse osquery Agent

## Installation

```bash
# Download the agent binary
wget https://github.com/rcourtman/Pulse/releases/latest/download/pulse-osquery-agent-linux-amd64

# Make it executable
chmod +x pulse-osquery-agent-linux-amd64

# Install osquery
wget https://pkg.osquery.io/deb/osquery_5.10.2-1.linux_amd64.deb
dpkg -i osquery_5.10.2-1.linux_amd64.deb
```

## Configuration

### Option 1: Configuration File (Recommended)

Create a configuration file:

```bash
cp osquery-agent.example.yaml /etc/pulse/osquery-agent.yaml
nano /etc/pulse/osquery-agent.yaml
```

Edit the configuration:

```yaml
server:
  url: "http://your-pulse-server:7655"
  api_token: "your-api-token"

agent:
  id: "my-server"
  interval: 60s

filter:
  mode: "aggressive"

logging:
  debug: false
```

Run with config file:

```bash
./pulse-osquery-agent-linux-amd64 --config /etc/pulse/osquery-agent.yaml
```

### Option 2: Command Line Flags

```bash
./pulse-osquery-agent-linux-amd64 \
  --url http://your-pulse-server:7655 \
  --api-token "your-token" \
  --filter-mode aggressive \
  --interval 60s
```

## Filter Modes

- **none**: No filtering, reports all processes and services
- **basic**: Filters ~10 common system patterns
- **aggressive**: Filters ~130+ system patterns (recommended)

## Custom Patterns

Add custom exclude patterns:

**Config file:**
```yaml
filter:
  mode: "aggressive"
  exclude_patterns:
    - "myapp*"
    - "*test*"
    - "debug*"
```

**Command line:**
```bash
--filter-mode aggressive --exclude-patterns "myapp*,*test*,debug*"
```

## Systemd Service

Create `/etc/systemd/system/pulse-osquery-agent.service`:

```ini
[Unit]
Description=Pulse osquery Agent
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/pulse-osquery-agent-linux-amd64 --config /etc/pulse/osquery-agent.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
systemctl daemon-reload
systemctl enable pulse-osquery-agent
systemctl start pulse-osquery-agent
systemctl status pulse-osquery-agent
```

## Flags

```
--config string
    Path to configuration file (YAML)
--url string
    Pulse server URL (default "http://localhost:7655")
--api-token string
    API token for authentication
--agent-id string
    Agent ID (defaults to hostname)
--interval duration
    Collection interval (default 1m0s)
--filter-mode string
    Filter mode: none, basic, aggressive (default "none")
--exclude-patterns string
    Comma-separated patterns to exclude
--debug
    Enable debug logging
```

## Troubleshooting

View logs:
```bash
journalctl -u pulse-osquery-agent -f
```

Test configuration:
```bash
./pulse-osquery-agent-linux-amd64 --config /etc/pulse/osquery-agent.yaml --debug
```
