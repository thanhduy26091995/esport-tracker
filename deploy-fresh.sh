#!/usr/bin/env bash
# deploy-fresh.sh
# First-time deployment for Ubuntu 22.04 VPS
# Usage:
#   chmod +x deploy-fresh.sh
#   ./deploy-fresh.sh

set -euo pipefail

# ═══════════════════════════════════════════════
# CONFIG
# ═══════════════════════════════════════════════
APP_DIR="$(cd "$(dirname "$0")" && pwd)"
BACKEND_PORT=8080
GO_VERSION="1.22.5"
NODE_VERSION="20"
DOMAIN="fifa.sitenow.cloud"
SSL_CERT="/etc/nginx/ssl/fullchain.pem"
SSL_KEY="/etc/nginx/ssl/privatekey.pem"

# ═══════════════════════════════════════════════
# COLORS
# ═══════════════════════════════════════════════
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log()  { echo -e "${GREEN}[$(date +%T)]${NC} $*"; }
warn() { echo -e "${YELLOW}[$(date +%T)] WARN:${NC} $*"; }
fail() { echo -e "${RED}[$(date +%T)] ERROR:${NC} $*"; exit 1; }

step() {
  echo ""
  echo -e "${BLUE}──────────────────────────────────────${NC}"
  echo -e "${BLUE} $*${NC}"
  echo -e "${BLUE}──────────────────────────────────────${NC}"
}

# ═══════════════════════════════════════════════
# PRECHECK
# ═══════════════════════════════════════════════
[[ "$EUID" -eq 0 ]] && fail "Run as sudo-capable user, not root."

[[ ! -d "$APP_DIR/backend" ]] && fail "Missing backend folder."
[[ ! -d "$APP_DIR/frontend" ]] && fail "Missing frontend folder."

SERVER_IP=$(curl -s ifconfig.me || hostname -I | awk '{print $1}')
HOST=${DOMAIN:-$SERVER_IP}

export PATH="/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:$PATH"
APT="sudo apt-get"

# ═══════════════════════════════════════════════
# 1. SYSTEM PACKAGES
# ═══════════════════════════════════════════════
step "1/9 — Install system packages"

$APT update -qq
$APT install -y -qq \
  curl git nginx ca-certificates gnupg lsb-release \
  build-essential unzip

log "System packages installed."

# ═══════════════════════════════════════════════
# 2. DOCKER
# ═══════════════════════════════════════════════
step "2/9 — Docker"

if ! command -v docker >/dev/null 2>&1; then
  curl -fsSL https://get.docker.com | sudo sh
  sudo usermod -aG docker "$USER"
  warn "Re-login may be required for docker group."
fi

if ! docker compose version >/dev/null 2>&1; then
  $APT install -y docker-compose-plugin
fi

log "Docker ready."

# ═══════════════════════════════════════════════
# 3. GO
# ═══════════════════════════════════════════════
step "3/9 — Go ${GO_VERSION}"

if ! command -v go >/dev/null 2>&1; then
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tar.gz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go.tar.gz
  rm -f /tmp/go.tar.gz

  echo 'export PATH=$PATH:/usr/local/go/bin' | sudo tee /etc/profile.d/go.sh >/dev/null
fi

export PATH="$PATH:/usr/local/go/bin"

log "$(go version)"

# ═══════════════════════════════════════════════
# 4. NODEJS
# ═══════════════════════════════════════════════
step "4/9 — Node ${NODE_VERSION}"

if ! command -v node >/dev/null 2>&1; then
  curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | sudo -E bash -
  $APT install -y nodejs
fi

log "Node: $(node -v)"
log "NPM : $(npm -v)"

# ═══════════════════════════════════════════════
# 5. ENV FILES
# ═══════════════════════════════════════════════
step "5/9 — Environment files"

BACKEND_ENV="$APP_DIR/backend/.env"
FRONTEND_ENV="$APP_DIR/frontend/.env"

[[ ! -f "$BACKEND_ENV" ]] && fail "backend/.env missing"

# fix BOM + CRLF
sed -i '1s/^\xEF\xBB\xBF//' "$BACKEND_ENV"
sed -i 's/\r$//' "$BACKEND_ENV"

log "backend/.env normalized"

cat > "$FRONTEND_ENV" <<EOF
VITE_API_BASE_URL=https://${DOMAIN}/api/v1
EOF

log "frontend/.env ready"

# load backend env safely
set -a
source "$BACKEND_ENV"
set +a

[[ -z "${DB_USER:-}" ]] && fail "DB_USER missing"
[[ -z "${DB_PASSWORD:-}" ]] && fail "DB_PASSWORD missing"
[[ -z "${DB_NAME:-}" ]] && fail "DB_NAME missing"

# ═══════════════════════════════════════════════
# 6. POSTGRES DOCKER
# ═══════════════════════════════════════════════
step "6/9 — PostgreSQL"

docker rm -f esport-postgres >/dev/null 2>&1 || true

docker run -d \
  --name esport-postgres \
  --restart unless-stopped \
  -e POSTGRES_USER="$DB_USER" \
  -e POSTGRES_PASSWORD="$DB_PASSWORD" \
  -e POSTGRES_DB="$DB_NAME" \
  -p 5432:5432 \
  -v esport_postgres_data:/var/lib/postgresql/data \
  postgres:14

log "Waiting PostgreSQL..."

for i in $(seq 1 60); do
  docker exec esport-postgres pg_isready -U "$DB_USER" >/dev/null 2>&1 && break
  sleep 1
done

log "PostgreSQL ready."

# ═══════════════════════════════════════════════
# 7. BUILD BACKEND
# ═══════════════════════════════════════════════
step "7/9 — Build backend"

cd "$APP_DIR/backend"

go mod tidy
go mod download
go build -o esport-backend ./cmd/server/

log "Backend built."

# ═══════════════════════════════════════════════
# 8. BUILD FRONTEND
# ═══════════════════════════════════════════════
step "8/9 — Build frontend"

cd "$APP_DIR/frontend"

npm install --silent
npm run build

log "Frontend built."
chmod o+x "$HOME"  # allow nginx to traverse home dir

# ═══════════════════════════════════════════════
# 9. SYSTEMD + NGINX
# ═══════════════════════════════════════════════
step "9/9 — Systemd + Nginx"

sudo tee /etc/systemd/system/esport-backend.service >/dev/null <<EOF
[Unit]
Description=Esport Backend
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

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable esport-backend
sudo fuser -k ${BACKEND_PORT}/tcp 2>/dev/null || true
sudo systemctl restart esport-backend

sleep 3

systemctl is-active --quiet esport-backend || fail "Backend failed"

sudo tee /etc/nginx/sites-available/esport >/dev/null <<EOF
# HTTP → HTTPS redirect
server {
    listen 80;
    server_name ${DOMAIN};
    return 301 https://\$host\$request_uri;
}

# HTTPS
server {
    listen 443 ssl;
    server_name ${DOMAIN};

    ssl_certificate     ${SSL_CERT};
    ssl_certificate_key ${SSL_KEY};
    ssl_protocols       TLSv1.2 TLSv1.3;
    ssl_ciphers         HIGH:!aNULL:!MD5;

    root ${APP_DIR}/frontend/dist;
    index index.html;

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://127.0.0.1:${BACKEND_PORT};
        proxy_http_version 1.1;
        proxy_set_header Host              \$host;
        proxy_set_header X-Real-IP         \$remote_addr;
        proxy_set_header X-Forwarded-For   \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }
}
EOF

sudo ln -sf /etc/nginx/sites-available/esport /etc/nginx/sites-enabled/esport
sudo rm -f /etc/nginx/sites-enabled/default

sudo nginx -t
sudo systemctl restart nginx

# ═══════════════════════════════════════════════
# DONE
# ═══════════════════════════════════════════════
echo ""
log "═══════════════════════════════════════════════"
log "DEPLOY SUCCESS"
log ""
log "App    : https://${DOMAIN}"
log "Health : https://${DOMAIN}/health"
log ""
log "Logs:"
log "sudo journalctl -u esport-backend -f"
log "docker logs esport-postgres"
log "═══════════════════════════════════════════════"