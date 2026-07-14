---
phase: testing
title: CI/CD Deployment Pipeline — Testing Strategy
description: Verification checklist for the GitHub Actions deployment pipeline
---

# Testing Strategy

## Test Coverage Goals

This feature has no unit-testable code (it's a YAML workflow + shell commands). Testing is entirely manual/integration:
- End-to-end: workflow triggers → SSH connects → deploy completes
- Security: no password in any log, no password needed by any team member
- Failure modes: workflow correctly fails and reports on bad deploys

## Unit Tests

N/A — no application code changed.

## Integration Tests

- [ ] **I1: SSH key auth works** — `ssh -i deploy_key deploy@<VPS_HOST> "echo OK"` succeeds from local machine
- [ ] **I2: deploy user can run deploy script** — `ssh -i deploy_key deploy@<VPS_HOST> "cd $APP_PATH && bash scripts/deploy-update.sh"` completes without permission errors
- [ ] **I3: deploy user can git pull** — `ssh -i deploy_key deploy@<VPS_HOST> "cd $APP_PATH && git fetch origin && git status"` works (no auth prompts)
- [ ] **I4: workflow triggers correctly** — GitHub Actions run appears in Actions tab after pressing "Run workflow"
- [ ] **I5: workflow log shows deploy output** — stdout from `deploy-update.sh` visible in GitHub Actions log
- [ ] **I6: wrong confirm input blocks deploy** — entering anything other than "deploy" in the confirm field causes job to be skipped

## End-to-End Tests

- [ ] **E1: Full deploy by VPS owner** — trigger via GitHub UI, verify app is updated on VPS
- [ ] **E2: Full deploy by team member (non-VPS-owner)** — member with repo access but no VPS password successfully triggers deploy
- [ ] **E3: Post-deploy smoke test** — app is accessible at production URL after deploy completes
- [ ] **E4: Deploy failure is visible** — intentionally break `deploy-update.sh` (e.g., wrong path), verify GitHub Actions run shows red/failed status with readable error

## Manual Testing

**Security checklist:**
- [ ] VPS password does NOT appear anywhere in GitHub Actions logs
- [ ] SSH private key does NOT appear in logs (GitHub should mask it, but verify)
- [ ] Team member can trigger deploy without being told the VPS password
- [ ] Removing a member's GitHub repo access prevents them from triggering future deploys

**VPS state checklist after deploy:**
- [ ] `docker ps` shows expected containers running
- [ ] App responds correctly at production URL
- [ ] `git log -1` on VPS shows expected commit SHA matching what was on `main`

## Performance Testing

Not applicable. Deploy time is the existing `deploy-update.sh` duration — this feature adds no overhead beyond SSH connection time (~1–2 seconds).

## Bug Tracking

**Common failure modes to watch for:**

| Symptom | Likely cause | Fix |
|---|---|---|
| `Permission denied (publickey)` | Public key not in VPS `authorized_keys` or wrong user | Re-check T1.2/T1.3 in planning |
| `Host key verification failed` | `known_hosts` not seeded | Add `ssh-keyscan` step to workflow |
| `git pull` auth error on VPS | Repo is private, deploy user has no GitHub access | Add VPS git deploy key to GitHub repo |
| `Permission denied` running docker | deploy user not in `docker` group | `usermod -aG docker deploy` + re-login |
| Workflow skipped (not failed) | `confirm` input wasn't exactly "deploy" | Re-run with correct input |
| App not updated after "success" | Script ran but deploy-update.sh is a no-op | Check script logic separately |
