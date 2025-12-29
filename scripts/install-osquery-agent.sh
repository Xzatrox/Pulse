#!/bin/bash
# Pulse osquery Agent Installation Script for Linux
set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Configuration
PULSE_URL="${PULSE_URL:-}"
PULSE_TOKEN="${PULSE_TOKEN:-}"
PULSE_AGENT_ID="${PULSE_AGENT_ID:-}"
PULSE_INTERVAL="${PULSE_INTERVAL:-60s}"
PULSE_FILTER_MODE="${PULSE_OSQUERY_FILTER_MODE:-aggressive}"
INSTALL_DIR="/usr/local/bin"
CONFIG_DIR="/etc/pulse"
SERVICE_NAME="pulse-osquery-agent"

# Functions
info() { echo -e "${CYAN}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[OK]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARN]${NC} $1"; }
error() { echo -e "${RED}[FAIL]${NC} $1"; exit 1; }

# Check if running as root
if [ "$EUID" -ne 0 ]; then
  error "This script must be run as root (use sudo)"
fi

# Detect architecture
detect_arch() {
  local arch=$(uname -m)
  case $arch in
    x86_64|amd64) echo "amd64" ;;
    aarch64|arm64) echo "arm64" ;;
    armv7l) echo "armv7" ;;
    armv6l) echo "armv6" ;;
    i386|i686) echo "386" ;;
    *) error "Unsupported architecture: $arch" ;;
  esac
}

ARCH=$(detect_arch)

# Banner
echo ""
echo "==========================================================="
echo "  Pulse osquery Agent - Linux Installation"
echo "==========================================================="
echo ""

# Interactive prompts if not set
if [ -z "$PULSE_URL" ]; then
  read -p "Enter Pulse server URL (e.g., http://pulse.example.com:7655): " PULSE_URL
fi

if [ -z "$PULSE_TOKEN" ]; then
  read -p "Enter API token: " PULSE_TOKEN
fi

if [ -z "$PULSE_TOKEN" ]; then
  error "API token is required"
fi

info "Configuration:"
echo "  Pulse URL: $PULSE_URL"
echo "  Token: ***${PULSE_TOKEN: -4}"
echo "  Agent ID: ${PULSE_AGENT_ID:-auto}"
echo "  Interval: $PULSE_INTERVAL"
echo "  Filter Mode: $PULSE_FILTER_MODE"
echo "  Architecture: $ARCH"
echo ""

# Install osquery if not present
if ! command -v osqueryi &> /dev/null; then
  info "Installing osquery..."
  if [ -f /etc/debian_version ]; then
    wget -q https://pkg.osquery.io/deb/osquery_5.10.2-1.linux_${ARCH}.deb
    dpkg -i osquery_5.10.2-1.linux_${ARCH}.deb || apt-get install -f -y
    rm osquery_5.10.2-1.linux_${ARCH}.deb
  elif [ -f /etc/redhat-release ]; then
    yum install -y https://pkg.osquery.io/rpm/osquery-5.10.2-1.linux.${ARCH}.rpm
  else
    warn "Could not auto-install osquery. Please install manually from https://osquery.io"
  fi
  success "osquery installed"
else
  success "osquery already installed"
fi

# Download agent
info "Downloading agent binary..."
DOWNLOAD_URL="${PULSE_URL}/download/pulse-osquery-agent?platform=linux&arch=${ARCH}"
curl -fsSL "$DOWNLOAD_URL" -o /tmp/pulse-osquery-agent || error "Failed to download agent"

# Verify checksum
info "Verifying checksum..."
CHECKSUM_URL="${PULSE_URL}/download/pulse-osquery-agent.sha256?platform=linux&arch=${ARCH}"
EXPECTED_CHECKSUM=$(curl -fsSL "$CHECKSUM_URL" | awk '{print $1}')
ACTUAL_CHECKSUM=$(sha256sum /tmp/pulse-osquery-agent | awk '{print $1}')

if [ "$EXPECTED_CHECKSUM" != "$ACTUAL_CHECKSUM" ]; then
  error "Checksum mismatch! Expected: $EXPECTED_CHECKSUM, Got: $ACTUAL_CHECKSUM"
fi
success "Checksum verified"

# Install binary
mv /tmp/pulse-osquery-agent "$INSTALL_DIR/pulse-osquery-agent"
chmod +x "$INSTALL_DIR/pulse-osquery-agent"
success "Installed to $INSTALL_DIR/pulse-osquery-agent"

# Create config directory
mkdir -p "$CONFIG_DIR"

# Generate agent ID from hostname if not provided
if [ -z "$PULSE_AGENT_ID" ]; then
  PULSE_AGENT_ID=$(hostname)
fi

# Validate filter mode
case "$PULSE_FILTER_MODE" in
  none|basic|aggressive)
    ;; # Valid modes
  *)
    warn "Unknown filter mode '$PULSE_FILTER_MODE', using 'basic'"
    PULSE_FILTER_MODE="basic"
    ;;
esac

# Create config file
cat > "$CONFIG_DIR/osquery-agent.yaml" <<EOF
server:
  url: "$PULSE_URL"
  api_token: "$PULSE_TOKEN"

agent:
  id: "${PULSE_AGENT_ID}"
  interval: ${PULSE_INTERVAL}

filter:
  mode: "$PULSE_FILTER_MODE"

logging:
  debug: false
EOF
success "Created configuration at $CONFIG_DIR/osquery-agent.yaml"

# Create systemd service
cat > "/etc/systemd/system/${SERVICE_NAME}.service" <<EOF
[Unit]
Description=Pulse osquery Agent
After=network.target

[Service]
Type=simple
ExecStart=$INSTALL_DIR/pulse-osquery-agent --config $CONFIG_DIR/osquery-agent.yaml
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and start service
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"
systemctl restart "$SERVICE_NAME"

# Wait and check status
sleep 3
if systemctl is-active --quiet "$SERVICE_NAME"; then
  success "Service started successfully!"
else
  warn "Service may not have started correctly"
  info "Check logs with: journalctl -u $SERVICE_NAME -f"
fi

echo ""
echo "==========================================================="
success "Installation complete!"
echo "==========================================================="
echo ""
info "Service Management:"
echo "  Start:   systemctl start $SERVICE_NAME"
echo "  Stop:    systemctl stop $SERVICE_NAME"
echo "  Restart: systemctl restart $SERVICE_NAME"
echo "  Status:  systemctl status $SERVICE_NAME"
echo "  Logs:    journalctl -u $SERVICE_NAME -f"
echo ""
info "Files installed:"
echo "  Binary: $INSTALL_DIR/pulse-osquery-agent"
echo "  Config: $CONFIG_DIR/osquery-agent.yaml"
echo "  Service: /etc/systemd/system/${SERVICE_NAME}.service"
echo ""
