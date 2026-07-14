---
phase: implementation
title: CI/CD Deployment Pipeline — Implementation Guide
description: Exact files and commands to implement GitHub Actions SSH-based deployment
---

# Implementation Guide

## Development Setup

**Prerequisites (VPS owner must complete Phase 1 from planning doc first):**
- SSH keypair generated locally
- Public key added to `deploy@51.79.249.111`'s `~/.ssh/authorized_keys`
- GitHub Secrets `VPS_HOST` and `VPS_SSH_KEY` added to repo

## Code Structure

```
.github/
  workflows/
    deploy.yml          ← new file (the entire feature)
scripts/
  deploy-update.sh      ← existing, unchanged
```

## Implementation Notes

### Core Feature: `.github/workflows/deploy.yml`

```yaml
name: Deploy to Production

on:
  workflow_dispatch:
    inputs:
      confirm:
        description: 'Type "deploy" to confirm'
        required: true
        default: ''
  # Optional: uncomment to also auto-deploy on every push to main
  # push:
  #   branches: [main]

jobs:
  deploy:
    name: Deploy to VPS
    runs-on: ubuntu-latest
    if: github.event_name == 'push' || github.event.inputs.confirm == 'deploy'

    steps:
      - name: Setup SSH
        run: |
          mkdir -p ~/.ssh
          echo "${{ secrets.VPS_SSH_KEY }}" > ~/.ssh/id_ed25519
          chmod 600 ~/.ssh/id_ed25519
          ssh-keyscan -H ${{ secrets.VPS_HOST }} >> ~/.ssh/known_hosts

      - name: Deploy
        run: |
          ssh deploy@${{ secrets.VPS_HOST }} \
            "cd /root/esport-tracker && \
             git fetch origin && \
             git reset --hard origin/main && \
             bash scripts/deploy-update.sh"
```

**Key notes:**
- `git reset --hard origin/main` instead of `git pull` — avoids merge conflicts if someone ever manually edited files on VPS
- `deploy` user at `51.79.249.111` already has sudo privileges for `systemctl` + `nginx` (required by `deploy-update.sh`)
- Repo is public — no additional GitHub deploy key needed for `git fetch`
- No third-party actions used — pure openssh, fully auditable

### VPS One-Time Commands

```bash
# ── On YOUR LOCAL machine ────────────────────────────────────
ssh-keygen -t ed25519 -C "github-actions-deploy-esport" -f ./deploy_key
# Produces: deploy_key (private) and deploy_key.pub (public)

# ── SSH into VPS with current password (last time) ───────────
ssh deploy@51.79.249.111

# ── On VPS: add the public key ───────────────────────────────
mkdir -p ~/.ssh
echo "<paste contents of deploy_key.pub here>" >> ~/.ssh/authorized_keys
chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys
exit

# ── Back on local: verify key auth works ─────────────────────
ssh -i ./deploy_key deploy@51.79.249.111 "echo SSH key auth OK"

# ── Dry run: test deploy command without committing ───────────
ssh -i ./deploy_key deploy@51.79.249.111 \
  "cd /root/esport-tracker && git fetch origin && git log --oneline -3"
```

After adding `VPS_SSH_KEY` to GitHub Secrets, delete the local private key:
```bash
rm ./deploy_key
# deploy_key.pub can be kept for reference
```

## Integration Points

**GitHub → VPS flow:**
1. Team member clicks "Run workflow" in GitHub Actions tab
2. Types `deploy` in the confirm field
3. GitHub Actions runner (ubuntu-latest, ephemeral) spins up
4. SSH key written to runner memory, used once, runner destroyed
5. Single SSH session: `git reset --hard` + `bash scripts/deploy-update.sh`
6. `deploy-update.sh` runs: Go build → systemctl restart → npm build (fifa + soc) → nginx reload
7. Exit code propagates to workflow pass/fail — visible in GitHub Actions tab

## Error Handling

- Any step returning non-zero → workflow fails → red X in GitHub Actions → team sees it immediately
- `git reset --hard origin/main` prevents stale state
- `deploy-update.sh` already has `set -eu` so any build/restart failure propagates correctly

## Security Notes

- **Never** print or echo the SSH key in any workflow step — GitHub masks secrets in logs but avoid it anyway
- The `confirm: "deploy"` guard prevents accidental workflow triggers
- Rotate the deploy SSH key: remove old key from VPS `~/.ssh/authorized_keys`, generate new keypair, update `VPS_SSH_KEY` secret
- Revoking a member's deploy access: remove their GitHub repo access (no VPS changes needed)
