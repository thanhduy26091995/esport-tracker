---
phase: design
title: CI/CD Deployment Pipeline — System Design
description: GitHub Actions workflow with SSH key authentication to replace manual VPS password deployment
---

# System Design & Architecture

## Architecture Overview

```mermaid
graph TD
    Dev["Team Member\n(GitHub UI)"] -->|workflow_dispatch — Run workflow button| GHA["GitHub Actions Runner\n(ubuntu-latest, ephemeral)"]
    GHA -->|SSH as root\nwith ed25519 key — NO password| VPS["VPS (Debian)\nroot@/root/esport-tracker"]
    VPS --> Pull["git fetch + git reset --hard origin/main"]
    Pull --> Script["bash scripts/deploy-update.sh"]
    Script --> Go["go build → esport-backend binary"]
    Script --> Systemctl["systemctl restart esport-backend"]
    Script --> NPM["npm run build (fifa + soc)"]
    Script --> Nginx["nginx reload"]

    subgraph "GitHub Secrets (encrypted)"
        S1["VPS_HOST — IP or hostname"]
        S2["VPS_SSH_KEY — private key (ed25519)"]
    end

    GHA -.reads.- S1
    GHA -.reads.- S2
```

**Key principle:** SSH private key lives only in GitHub Secrets (encrypted at rest). No password is ever stored or transmitted. VPS user is `root` because `deploy-update.sh` requires system-level operations (systemctl, nginx, tee to /etc/nginx) — a restricted deploy user would need equivalent sudo access anyway.

## Data Models

No application data model changes. This feature is infrastructure-only.

**GitHub Secrets schema:**

| Secret name | Value | Example |
|---|---|---|
| `VPS_HOST` | VPS IP or domain | `123.45.67.89` or `vps.example.com` |
| `VPS_SSH_KEY` | Private key (PEM/OpenSSH format) | `-----BEGIN OPENSSH PRIVATE KEY-----...` |

App path `/root/esport-tracker` and user `root` are hardcoded in the workflow (not secrets) since they never change.

## API Design

No new API endpoints. The "API" here is the GitHub Actions workflow file.

**Workflow trigger interface:**

```yaml
on:
  workflow_dispatch:          # Manual "button click" in GitHub UI
    inputs:
      confirm:
        description: 'Type "deploy" to confirm production deployment'
        required: true
  push:
    branches: [main]          # Optional: auto-deploy on push to main
```

The `confirm` input on `workflow_dispatch` acts as a lightweight safeguard against accidental clicks.

## Component Breakdown

### 1. GitHub Actions Workflow (`.github/workflows/deploy.yml`)
- Runs on `ubuntu-latest` (GitHub-hosted runner, free tier)
- Steps:
  1. Load SSH private key from secret into ssh-agent
  2. Add VPS host to `known_hosts` (prevent MITM prompt)
  3. SSH into VPS, `cd $VPS_APP_PATH && git pull origin main && bash scripts/deploy-update.sh`
  4. Report exit code as workflow success/failure

### 2. VPS — Deploy User Setup (one-time manual step)
- Create OS user `deploy` with no password login
- Clone (or `chown`) app directory to `deploy` user
- Add GitHub Actions public key to `~/.ssh/authorized_keys`
- Configure `sudoers` if `deploy-update.sh` requires root-level Docker commands

### 3. VPS — SSH Hardening (one-time manual step)
- `PasswordAuthentication no` in `/etc/ssh/sshd_config` for deploy user
- Keeps existing root/admin access via password untouched (owner only)

### 4. SSH Keypair (one-time generation)
- Generated locally: `ssh-keygen -t ed25519 -C "github-actions-deploy-esport" -f deploy_key`
- Public key → VPS `authorized_keys`
- Private key → GitHub Secret `VPS_SSH_KEY`
- Local key files deleted after setup

## Design Decisions

**Why GitHub Actions over Jenkins?**
- No additional server infrastructure required
- Secrets management is built-in, encrypted, and access-controlled via GitHub permissions
- "Button click" UX is native (`workflow_dispatch`)
- Revoking deploy access = removing GitHub repo access (no VPS changes needed)

**Why ed25519 over RSA?**
- Shorter key, faster, more modern and secure

**Why SSH as `root` instead of a dedicated `deploy` user?**
- `deploy-update.sh` requires `sudo systemctl restart esport-backend`, `sudo tee /etc/nginx/...`, and `sudo systemctl reload nginx`
- A dedicated deploy user would need NOPASSWD sudo for all these — effectively root-equivalent access
- App lives in `/root/esport-tracker`, chown to another user adds unnecessary complexity
- The primary security goal is **eliminating password sharing**, not limiting SSH user scope

**Why not GitHub Actions self-hosted runner on VPS?**
- Adds persistent process to maintain on VPS
- Overkill for this use case; SSH-based approach is simpler and equally secure

**Why keep `deploy-update.sh` unchanged?**
- Existing script is tested and working; this feature is purely about the trigger mechanism
- Changes to the script are a separate concern

## Non-Functional Requirements

**Security:**
- Private key never touches the filesystem unencrypted during the workflow (loaded into ssh-agent in memory)
- `StrictHostKeyChecking` enforced via pre-seeded `known_hosts` (prevents MITM on first connect)
- Deploy user has no sudo except what `deploy-update.sh` specifically requires

**Reliability:**
- Workflow fails loudly if any step returns non-zero exit code
- GitHub retains logs for 90 days by default

**Auditability:**
- GitHub Actions history shows: who triggered, when, which commit SHA was deployed, full stdout/stderr log
