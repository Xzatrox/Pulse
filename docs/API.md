# 🔌 Pulse API Reference

Pulse provides a comprehensive REST API for automation and integration.

**Base URL**: `http://<your-pulse-ip>:7655/api`

## 🔐 Authentication

All API requests require authentication via one of the following methods:

**1. API Token (Recommended)**
Pass the token in the `X-API-Token` header.
```bash
curl -H "X-API-Token: your-token" http://localhost:7655/api/health
```

**2. Bearer Token**
```bash
curl -H "Authorization: Bearer your-token" http://localhost:7655/api/health
```

**3. Session Cookie**
Standard browser session cookie (used by the UI).

---

## 📡 Core Endpoints

### System Health
`GET /api/health`
Check if Pulse is running.
```json
{ "status": "healthy", "uptime": 3600 }
```

### System State
`GET /api/state`
Returns the complete state of your infrastructure (Nodes, VMs, Containers, Storage, Alerts). This is the main endpoint used by the dashboard.

### Version Info
`GET /api/version`
Returns version, build time, and update status.

---

## 🖥️ Nodes & Config

### List Nodes
`GET /api/config/nodes`

### Add Node
`POST /api/config/nodes`
```json
{
  "type": "pve",
  "name": "Proxmox 1",
  "host": "https://192.168.1.10:8006",
  "user": "root@pam",
  "password": "password"
}
```

### Test Connection
`POST /api/config/nodes/test-connection`
Validate credentials before saving.

---

## 📊 Metrics & Charts

### Chart Data
`GET /api/charts?range=1h`
Returns time-series data for CPU, Memory, and Storage.
**Ranges**: `1h`, `24h`, `7d`, `30d`

### Storage Stats
`GET /api/storage/`
Detailed storage usage per node and pool.

### Backup History
`GET /api/backups/unified`
Combined view of PVE and PBS backups.

---

## 🔔 Notifications

### Send Test Notification
`POST /api/notifications/test`
Triggers a test alert to all configured channels.

### Manage Webhooks
- `GET /api/notifications/webhooks`
- `POST /api/notifications/webhooks`
- `DELETE /api/notifications/webhooks/<id>`

---

## 🛡️ Security

### List API Tokens
`GET /api/security/tokens`

### Create API Token
`POST /api/security/tokens`
```json
{ "name": "ansible-script", "scopes": ["monitoring:read"] }
```

### Revoke Token
`DELETE /api/security/tokens/<id>`

---

## ⚙️ System Settings

### Get Settings
`GET /api/system/settings`
Retrieve current system settings.

### Update Settings
`POST /api/system/settings/update`
Update system settings. Requires admin + `settings:write`.

### Toggle Mock Mode
`POST /api/system/mock-mode`
Enable or disable mock data generation (dev/demo only).

---

## 🔑 OIDC / SSO

### Get OIDC Config
`GET /api/security/oidc`
Retrieve current OIDC provider settings.

### Update OIDC Config
`POST /api/security/oidc`
Configure OIDC provider details (Issuer, Client ID, etc).

### Login
`GET /api/oidc/login`
Initiate OIDC login flow.

---

## 🤖 Pulse AI *(v5)*

### Get AI Settings
`GET /api/settings/ai`
Returns current AI configuration (providers, models, patrol status). Requires admin + `settings:read`.

### Update AI Settings
`PUT /api/settings/ai/update` (or `POST /api/settings/ai/update`)
Configure AI providers, API keys, and preferences. Requires admin + `settings:write`.

### List Models
`GET /api/ai/models`
Lists models available to the configured providers (queried live from provider APIs).

### Execute (Chat + Tools)
`POST /api/ai/execute`
Runs an AI request which may return tool calls, findings, or suggested actions.

### Execute (Streaming)
`POST /api/ai/execute/stream`
Streaming variant of execute (used by the UI for incremental responses).

### Patrol
- `GET /api/ai/patrol/status`
- `GET /api/ai/patrol/findings`
- `GET /api/ai/patrol/history`
- `POST /api/ai/patrol/run` (admin)

### Cost Tracking
`GET /api/ai/cost/summary`
Get AI usage statistics (includes retention window details).

## 📈 Metrics Store (v5)

### Store Stats
`GET /api/metrics-store/stats`
Returns stats for the persistent metrics store (SQLite-backed).

### History
`GET /api/metrics-store/history`
Returns historical metric series for a resource and time range.

---

## 🤖 Agent Endpoints

### Unified Agent (Recommended)
`GET /download/pulse-agent`
Downloads the unified agent binary for the current platform.

The unified agent combines host, Docker, and Kubernetes monitoring. Use `--enable-docker` or `--enable-kubernetes` to enable additional metrics.

See [UNIFIED_AGENT.md](UNIFIED_AGENT.md) for installation instructions.

### Unified Agent Installer Script
`GET /install.sh`
Serves the universal `install.sh` used to install `pulse-agent` on target machines.

### Legacy Agents (Deprecated)
`GET /download/pulse-host-agent` - *Deprecated, use pulse-agent*
`GET /download/pulse-docker-agent` - *Deprecated, use pulse-agent --enable-docker*

### Submit Reports
`POST /api/agents/host/report` - Host metrics
`POST /api/agents/docker/report` - Docker container metrics
`POST /api/agents/kubernetes/report` - Kubernetes cluster metrics

---

> **Note**: This is a summary of the most common endpoints. For a complete list, inspect the network traffic of the Pulse dashboard or check the source code in `internal/api/router.go`.


## Osquery Endpoints

### POST /api/agents/{agentID}/osquery
Submit osquery report from agent.

**Scope**: `host:report`

**Request Body**:
```json
{
  "processes": [{"pid": "1234", "name": "app", "path": "/usr/bin/app"}],
  "services": [{"name": "sshd", "state": "running", "status": "active"}],
  "timestamp": "2024-01-01T00:00:00Z"
}
```

**Response**: `{"success": true, "message": "Report received"}`

### GET /api/osquery/reports
Get all latest osquery reports from all agents.

**Scope**: `monitoring:read`

**Response**:
```json
{
  "agent-1": {
    "timestamp": "2024-01-01T00:00:00Z",
    "processes": [...],
    "services": [...]
  }
}
```
