---
phase: requirements
title: CI/CD Deployment Pipeline
description: Automate VPS deployment via GitHub Actions so any team member can deploy with one click without SSH password access
---

# Requirements & Problem Understanding

## Problem Statement

Currently, deploying the app to the VPS requires:
1. SSH-ing into the VPS with the root/admin password
2. Navigating to the app directory
3. Running `git pull` + `bash scripts/deploy-update.sh`
4. Entering the VPS password when prompted

**Pain points:**
- VPS password must be shared with every team member who needs to deploy — a significant security risk
- If any member's machine is compromised, the entire VPS is at risk
- Manual process is error-prone (wrong directory, wrong branch, wrong script)
- No audit trail of who deployed what and when

## Goals & Objectives

**Primary goals:**
- Any authorized team member can trigger a production deployment from GitHub UI with one button click
- No team member ever needs to know the VPS password
- Deployment is reproducible and auditable (who triggered it, when, which commit)

**Secondary goals:**
- Auto-deploy on push to `main` branch (optional, team decision)
- Deployment status visible in GitHub (pass/fail with logs)

**Non-goals:**
- Multiple environments (staging, dev) — out of scope, single production only
- Blue/green or zero-downtime deploys — existing `deploy-update.sh` behavior is kept as-is
- Containerized CI build pipeline — app is already Dockerized on VPS; we only orchestrate the trigger

## User Stories & Use Cases

- As a **team member**, I want to click "Run workflow" in GitHub Actions so that I can deploy the latest `main` branch without needing the VPS password.
- As a **team member**, I want to see the deployment log in GitHub so that I know if the deploy succeeded or failed.
- As the **VPS owner**, I want to revoke a member's deploy access without changing the VPS password so that offboarding is safe and isolated.
- As the **VPS owner**, I want the VPS to only accept SSH connections via key (not password) for the deploy user so that brute-force password attacks are mitigated.

## Success Criteria

- [ ] Deployment triggers successfully via GitHub Actions "Run workflow" button
- [ ] No VPS password is stored anywhere in the repo or GitHub Secrets
- [ ] VPS SSH password authentication is disabled for the deploy user (key-only)
- [ ] Deployment logs (stdout/stderr of `deploy-update.sh`) are visible in GitHub Actions run
- [ ] Revoking a member's GitHub repo access also revokes their ability to trigger deploys

## Constraints & Assumptions

- Repo is hosted on GitHub (public or private)
- VPS runs Debian/Linux with Docker already installed
- Existing `scripts/deploy-update.sh` works correctly and will not be modified
- The app repo is already cloned on the VPS at a known path (e.g., `/root/esport-tracker` or similar)
- Team uses `main` as the production branch
- GitHub Actions free tier minutes are sufficient (small project)

## Questions & Open Items

- [x] **App path on VPS**: `/root/esport-tracker` ✓
- [x] **Repo visibility**: Public — no GitHub deploy key needed for `git pull` ✓
- [x] **Runtime**: NOT Docker — native Go binary + `systemctl` + Nginx directly ✓
- [x] **deploy-update.sh sudo usage**: `sudo systemctl restart esport-backend`, `sudo tee /etc/nginx/...`, `sudo nginx -t && sudo systemctl reload nginx` ✓
- [x] **OS user**: SSH as `root` (app lives in `/root/`, script needs system-level sudo — dedicated deploy user would need full sudo anyway)
- [ ] Should deploy auto-trigger on push to `main`, or manual-only?
