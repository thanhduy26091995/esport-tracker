---
phase: planning
title: CI/CD Deployment Pipeline — Planning
description: Task breakdown for GitHub Actions + SSH key deployment setup
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1: VPS prepared** — deploy user created, SSH key auth configured
- [ ] **M2: GitHub configured** — secrets added, workflow file committed
- [ ] **M3: First successful deploy via button** — end-to-end verified

## Task Breakdown

### Phase 1: VPS Setup (done once by VPS owner, ~15 min)

- [ ] **T1.1** Generate SSH keypair locally (on your machine, NOT on the VPS)
  ```bash
  ssh-keygen -t ed25519 -C "github-actions-deploy-esport" -f ./deploy_key
  # produces: deploy_key (private) and deploy_key.pub (public)
  ```

- [ ] **T1.2** Add public key to deploy user's authorized_keys on VPS
  ```bash
  # SSH into VPS with your current password (last time you'll need it for this)
  ssh deploy@51.79.249.111

  # On VPS:
  mkdir -p ~/.ssh
  echo "<paste contents of deploy_key.pub here>" >> ~/.ssh/authorized_keys
  chmod 700 ~/.ssh && chmod 600 ~/.ssh/authorized_keys
  ```

- [ ] **T1.3** Test SSH login with deploy key from local machine (confirm before deleting key)
  ```bash
  ssh -i ./deploy_key deploy@51.79.249.111 "echo SSH key auth works"
  ```

- [ ] **T1.4** Test full deploy command via key (dry run)
  ```bash
  ssh -i ./deploy_key deploy@51.79.249.111 \
    "cd /root/esport-tracker && git fetch origin && git status"
  ```

- [ ] **T1.5** Delete local private key file after confirming it's in GitHub Secrets
  ```bash
  rm ./deploy_key  # deploy_key.pub can be kept for reference
  ```

### Phase 2: GitHub Configuration (~10 min)

- [ ] **T2.1** Add GitHub Secrets (repo Settings → Secrets and variables → Actions → New repository secret):
  - `VPS_HOST` → `51.79.249.111`
  - `VPS_SSH_KEY` → full contents of `./deploy_key` (private key, including the `-----BEGIN OPENSSH PRIVATE KEY-----` header/footer)

- [ ] **T2.2** Create workflow file `.github/workflows/deploy.yml` (see implementation doc)

- [ ] **T2.3** Commit and push workflow file to `main`

### Phase 3: Verification (~15 min)

- [ ] **T3.1** Trigger first deploy via GitHub UI (Actions → Deploy to Production → Run workflow)
- [ ] **T3.2** Verify deployment log shows git pull output and deploy-update.sh output
- [ ] **T3.3** Verify app is running correctly on VPS after deploy
- [ ] **T3.4** Have a second team member trigger a deploy to confirm access works without VPS password

### Phase 4: Optional Hardening

- [ ] **T4.1** Disable password SSH for deploy user in `/etc/ssh/sshd_config`
  ```
  Match User deploy
      PasswordAuthentication no
  ```
- [ ] **T4.2** Add `known_hosts` fingerprint to workflow (prevent first-connect prompt)
- [ ] **T4.3** Add Slack/email notification on deploy failure (via GitHub Actions notification step)

## Dependencies

- T1.4 depends on knowing the exact app path on VPS (open question from requirements)
- T2.2 depends on T1.1–T1.6 completing successfully (need to know the deploy user name)
- T3.1 depends on T2.1–T2.3 all done

## Timeline & Estimates

| Phase | Effort | Who |
|---|---|---|
| Phase 1: VPS Setup | ~30 min | VPS owner (1 person) |
| Phase 2: GitHub Config | ~10 min | Any team member with repo admin |
| Phase 3: Verification | ~15 min | VPS owner + 1 other member |
| Phase 4: Hardening | ~20 min | VPS owner (optional) |
| **Total** | **~1 hour** | |

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| `deploy-update.sh` requires interactive password prompt | Audit the script before proceeding; replace any `sudo` password prompts with sudoers NOPASSWD rules |
| SSH key leaks via GitHub (e.g., accidental log print) | Key is in secret; workflow must not `echo $VPS_SSH_KEY` — review workflow carefully |
| VPS known_hosts not pre-seeded → workflow hangs on prompt | Pre-seed with `ssh-keyscan` in workflow or hardcode fingerprint in known_hosts step |
| `git pull` fails due to local uncommitted changes on VPS | Ensure VPS repo is always clean (no manual edits); add `git fetch && git reset --hard origin/main` to be safe |
| deploy user lacks permission to restart Docker services | Add to `docker` group (T1.5) before testing |

## Resources Needed

- VPS root/admin access (one-time setup only)
- GitHub repo admin access (to add Secrets)
- Contents of `scripts/deploy-update.sh` (to verify it works non-interactively as deploy user)
