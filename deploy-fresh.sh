#!/usr/bin/env bash
# deploy-fresh.sh — First-time setup on a fresh Ubuntu 22.04 VPS
#
# Usage:
#   1. Copy your code to the VPS at APP_DIR (e.g. via scp/rsync/git)
#   2. Edit the CONFIG section below
#   3. bash deploy-fresh.sh

set -eu

# ══════════════════════════════════════════════════════════════
# CONFIG — edit these before running
# ══════════════════════════════════════════════════════════════
APP_DIR="/opt/esport"
BACKEND_PORT=8080
GO_VERSION="1.22.5"
NODE_VERSION="20"

# Domain or public IP for Nginx server_name and VITE_API_BASE_URL
# Leave empty to auto-detect the server's public IP
DOMAIN=""

# Database credentials (used for Docker container + backend .env)
DB_USER="postgres"
DB_PASSWORD="changeme"          # ← change this
DB_NAME="esport_tracker"
# ══════════════════════════════════════════════════════════════

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log()  { echo -e "${GREEN}[$(date +%T)]${NC} $*"; }
warn() { echo -e "${YELLOW}[$(date +%T)] WARN:${NC} $*"; }
fail() { echo -e "${RED}[$(date +%T)] ERROR:${NC} $*"; exit 1; }
step() { echo -e "\n${BLUE}──────────────────────────────────────${NC}"; echo -e "${BLUE} $*${NC}"; echo -e "${BLUE}──────────────────────────────────────${NC}"; }

[[ "$EUID" -eq 0 ]] && fail "Run as a sudo-capable user, not root."
[[ ! -d "$APP_DIR/backend" ]] && fail "Code not found at $APP_DIR. Copy your project there first."

SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')
HOST=${DOMAIN:-$SERVER_IP}

# ── 1. System packages ────────────────────────────────────────
step "1/9 — System packages"
sudo apt-get update -qq
sudo apt-get install -y -qq \
  curl git nginx ca-certificates gnupg lsb-release \
  build-essential
log "System packages installed."

# ── 2. Docker ─────────────────────────────────────────────────
step "2/9 — Docker"
if ! command -v docker &>/dev/null; then
  log "Installing Docker..."
  curl -fsSL https://get.docker.com | sudo sh
  sudo usermod -aG docker "$USER"
  warn "Added $USER to the docker group. If docker commands fail, run: newgrp docker"
  exec sg docker "$0 $*" 2>/dev/null || true
else
  log "Docker already installed: $(docker --version)"
fi

if ! docker compose version &>/dev/null 2>&1; then
  sudo apt-get install -y docker-compose-plugin
fi

# ── 3. Go ─────────────────────────────────────────────────────
step "3/9 — Go $GO_VERSION"
if ! command -v go &>/dev/null; then
  log "Installing Go $GO_VERSION ..."
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go.tar.gz
  rm /tmp/go.tar.gz
  echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh > /dev/null
  export PATH="$PATH:/usr/local/go/bin"
  log "Go installed: $(go version)"
else
  export PATH="$PATH:/usr/local/go/bin"
  log "Go already installed: $(go version)"
fi

# ── 4. Node.js ────────────────────────────────────────────────
step "4/9 — Node.js $NODE_VERSION"
if ! command -v node &>/dev/null; then
  log "Installing Node.js $NODE_VERSION ..."
  curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | sudo -E bash - -qq
  sudo apt-get install -y -qq nodejs
  log "Node installed: $(node --version)"
else
  log "Node already installed: $(node --version)"
fi

# ── 5. Environment files ──────────────────────────────────────
step "5/9 — Environment files"

BACKEND_ENV="$APP_DIR/backend/.env"
if [ ! -f "$BACKEND_ENV" ]; then
  cat > "$BACKEND_ENV" <<EOF
DB_HOST=localhost
DB_PORT=5432
DB_USER=${DB_USER}
DB_PASSWORD=${DB_PASSWORD}
DB_NAME=${DB_NAME}
DB_SSLMODE=disable
PORT=${BACKEND_PORT}
CORS_ORIGINS=http://${HOST},https://${HOST}
EOF
  log "Created $BACKEND_ENV"
else
  warn "$BACKEND_ENV already exists — skipping. Review it manually."
fi

FRONTEND_ENV="$APP_DIR/frontend/.env"
if [ ! -f "$FRONTEND_ENV" ]; then
  cat > "$FRONTEND_ENV" <<EOF
VITE_API_BASE_URL=http://${HOST}/api/v1
EOF
  log "Created $FRONTEND_ENV"
else
  warn "$FRONTEND_ENV already exists — skipping."
fi

# ── 6. PostgreSQL via Docker ──────────────────────────────────
step "6/9 — PostgreSQL (Docker)"
cd "$APP_DIR/backend"

sed -i \
  -e "s/POSTGRES_PASSWORD: postgres/POSTGRES_PASSWORD: ${DB_PASSWORD}/" \
  -e "s/POSTGRES_USER: postgres/POSTGRES_USER: ${DB_USER}/" \
  -e "s/POSTGRES_DB: esport_tracker/POSTGRES_DB: ${DB_NAME}/" \
  docker-compose.yml

if docker ps --format '{{.Names}}' | grep -q "esport-postgres"; then
  log "PostgreSQL container already running."
else
  docker compose up -d postgres
  log "Waiting for PostgreSQL to be ready..."
  for i in $(seq 1 30); do
    docker exec esport-postgres pg_isready -U "$DB_USER" &>/dev/null && break
    sleep 1
    [[ $i -eq 30 ]] && fail "PostgreSQL did not become ready in time."
  done
  log "PostgreSQL is ready."
fi

# ── 7. Build backend ──────────────────────────────────────────
step "7/9 — Build Go backend"
cd "$APP_DIR/backend"
go mod download
go build -o esport-backend ./cmd/server/
log "Binary: $APP_DIR/backend/esport-backend"

# ── 8. Build frontend ─────────────────────────────────────────
step "8/9 — Build frontend"
cd "$APP_DIR/frontend"
npm ci --silent
npm run build
log "Frontend built to $APP_DIR/frontend/dist/"

# ── 9. Systemd + Nginx ────────────────────────────────────────
step "9/9 — Systemd + Nginx"

log "Creating systemd service..."
sudo tee /etc/systemd/system/esport-backend.service > /dev/null <<EOF
[Unit]
Description=FC25 Esport Score Tracker — Go backend
After=network.target docker.service
Wants=docker.service

[Service]
Type=simple
User=${USER}
WorkingDirectory=${APP_DIR}/backend
EnvironmentFile=${APP_DIR}/backend/.env
ExecStart=${APP_DIR}/backend/esport-backend
Restart=always
RestartSec=5
StandardOutput=journal
StandardError=journal

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable esport-backend
sudo systemctl start esport-backend

sleep 2
if ! systemctl is-active --quiet esport-backend; then
  fail "Backend service failed to start. Debug: sudo journalctl -u esport-backend -n 50"
fi
log "Backend service running."

log "Configuring Nginx..."
sudo tee /etc/nginx/sites-available/esport > /dev/null <<EOF
server {
    listen 80;
    server_name ${DOMAIN:-_};
    client_max_body_size 10M;

    root ${APP_DIR}/frontend/dist;
    index index.html;

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    location /api/ {
        proxy_pass         http://127.0.0.1:${BACKEND_PORT};
        proxy_http_version 1.1;
        proxy_set_header   Host              \$host;
        proxy_set_header   X-Real-IP         \$remote_addr;
        proxy_set_header   X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header   X-Forwarded-Proto \$scheme;
        proxy_read_timeout 60s;
    }
}
EOF

sudo ln -sf /etc/nginx/sites-available/esport /etc/nginx/sites-enabled/esport
sudo rm -f /etc/nginx/sites-enabled/default
sudo nginx -t
sudo systemctl enable nginx
sudo systemctl reload nginx
log "Nginx configured."

# ── Done ──────────────────────────────────────────────────────
echo ""
log "════════════════════════════════════════"
log " Fresh deploy complete!"
log ""
log " App    →  http://${HOST}"
log " Health →  http://${HOST}/api/v1/health"
log ""
log " Logs   →  sudo journalctl -u esport-backend -f"
log " DB     →  docker logs esport-postgres"
log "════════════════════════════════════════"
