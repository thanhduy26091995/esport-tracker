#!/usr/bin/env bash
# deploy-update.sh — Re-deploy after updating code on the VPS
#
# Usage:
#   bash deploy-update.sh   ← must be bash, not sh

set -eu

# ══════════════════════════════════════════════════════════════
# CONFIG — must match deploy-fresh.sh
# ══════════════════════════════════════════════════════════════
APP_DIR="$(cd "$(dirname "$0")/.." && pwd)"
BACKEND_PORT=8080
DOMAIN="fifa.sitenow.cloud"
SSL_CERT="/etc/nginx/ssl/fullchain.pem"
SSL_KEY="/etc/nginx/ssl/privatekey.pem"
# ══════════════════════════════════════════════════════════════

RED='\033[0;31m'; GREEN='\033[0;32m'; YELLOW='\033[1;33m'; BLUE='\033[0;34m'; NC='\033[0m'
log()  { echo -e "${GREEN}[$(date +%T)]${NC} $*"; }
warn() { echo -e "${YELLOW}[$(date +%T)] WARN:${NC} $*"; }
fail() { echo -e "${RED}[$(date +%T)] ERROR:${NC} $*"; exit 1; }
step() { echo -e "\n${BLUE}──────────────────────────────────────${NC}"; echo -e "${BLUE} $*${NC}"; echo -e "${BLUE}──────────────────────────────────────${NC}"; }

export PATH="$PATH:/usr/local/go/bin"

[[ ! -d "$APP_DIR/backend" ]] && fail "App not found at $APP_DIR. Run deploy-fresh.sh first."

# ── 1. Rebuild backend ────────────────────────────────────────
step "1/4 — Rebuild Go backend"
cd "$APP_DIR/backend"
go mod download
go build -o esport-backend ./cmd/server/
log "Backend binary rebuilt."

# ── 2. Restart backend service ───────────────────────────────
step "2/4 — Restart backend service"
sudo systemctl restart esport-backend
sleep 2
if systemctl is-active --quiet esport-backend; then
  log "Backend service restarted successfully."
else
  fail "Backend failed to restart. Debug: sudo journalctl -u esport-backend -n 50"
fi

# ── 3. Rebuild frontend ───────────────────────────────────────
step "3/4 — Rebuild frontend"
cd "$APP_DIR/frontend"
npm install --silent
npm run build
log "Frontend rebuilt."

# ── 4. Update Nginx config + Reload ──────────────────────────
step "4/4 — Update Nginx config and reload"
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

    client_max_body_size 5M;

    root ${APP_DIR}/frontend/dist;
    index index.html;

    location / {
        try_files \$uri \$uri/ /index.html;
    }

    location /uploads/ {
        alias ${APP_DIR}/backend/uploads/;
        expires 7d;
        add_header Cache-Control "public, immutable";
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
sudo nginx -t && sudo systemctl reload nginx
log "Nginx config updated and reloaded."

# ── Health check ──────────────────────────────────────────────
SERVER_IP=$(curl -s ifconfig.me 2>/dev/null || hostname -I | awk '{print $1}')
HTTP_STATUS=$(curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:${BACKEND_PORT}/health" || echo "000")

echo ""
log "════════════════════════════════════════"
log " Update deploy complete!"
log " Health : /api/v1/health → HTTP $HTTP_STATUS"
log " Site   : http://${SERVER_IP}"
log ""
if [[ "$HTTP_STATUS" != "200" ]]; then
  warn " Backend health check returned $HTTP_STATUS"
  warn " Check logs: sudo journalctl -u esport-backend -n 50"
fi
log "════════════════════════════════════════"
