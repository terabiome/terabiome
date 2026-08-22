# HFC setup and hardening checklist

Status legend: `[ ]` not started, `[-]` in progress, `[x]` verified, `[!]`
blocked or needs a decision.

This is a review-first checklist. Do not make a security-changing action until
the current state, verification method, and recovery path are understood.

## 0. Guardrails and access recovery

- [ ] Confirm the exact server is the intended HostHatch HFC instance.
- [ ] Record the provider console/recovery access method outside this repo.
- [x] Confirm the SSH login user and key-based access path.
- [ ] Confirm a second recovery
  session before changing SSH or firewall settings.
- [x] Record the OS and release: Debian GNU/Linux 13.6 (trixie).
- [ ] Record the current SSH host key fingerprint outside this repo and verify
  it on future connections.

## 1. Read-only baseline assessment

- [x] Inspect identity: hostname, OS/release, kernel, uptime, timezone, and
  configured time synchronization.
- [-] Inspect compute and disk: CPU, RAM, filesystem capacity, mount points,
  block devices, and swap.
- [x] Verify virtualization: KVM virtual machine (`systemd-detect-virt`).
- [x] Inspect access: local users, sudo policy, SSH daemon configuration, and
  authorized-key locations.
- [x] Inspect exposure: listening ports, firewall state, public network
  interfaces, and active services/timers.
- [ ] Inspect maintenance posture: pending updates, automatic-update settings,
  recent failed logins, and system logs.
- [ ] Save a sanitized baseline summary in this directory; exclude IPs, keys,
  host fingerprints, and other secrets.

## 2. Establish the administrative baseline

- [ ] Decide the administrator account model and ensure it has key-only SSH
  access and least-privilege sudo.
- [ ] Ensure a second authorized administrator/recovery path exists before
  disabling any existing access path.
- [ ] Apply current security updates after reviewing the package changes.
- [ ] Set correct timezone and reliable time synchronization.
- [ ] Create an update and reboot policy appropriate to an offline research box.

## 3. Harden remote access and network exposure

- [x] Enroll the host in the existing Tailscale tailnet and verify private SSH.
- [x] Install and enable UFW with default-deny inbound policy; permit Tailscale
  interface traffic and its required UDP transport.
- [x] Restrict Cockpit to private Tailscale access; public TCP port 9090 is
  blocked.
- [ ] Restrict SSH to key authentication; disable password and root SSH login
  only after recovery access is tested.
- [ ] Set conservative SSH session and forwarding policies based on actual
  administration needs.
- [ ] Configure a default-deny host firewall, allowing only necessary SSH
  access. Keep a live recovery session open while testing.
- [ ] Remove, stop, or disable services not required for the research-box role.
- [ ] Verify externally and locally that no unintended services are reachable.

## 4. Storage and Flora data layout

- [ ] Verify filesystem type, free space, mount behavior, and disk-health
  visibility; decide whether any storage changes are required.
- [x] Confirm `/var/data/flora` is the existing Flora data root.
- [ ] Verify ownership, permissions, and mount expectations for
  `/var/data/flora`.
- [ ] Create the initial data-area structure for immutable raw data, normalized
  Parquet, derived features, labels, temporary work, and logs.
- [ ] Define write permissions so raw vendor data cannot be casually modified
  or deleted by routine research processes.
- [ ] Define local retention and capacity monitoring for rebuildable working
  data.
- [ ] Verify the R2 raw-data source and test the HFC rebuild path from it.

## 5. Research host foundations

- [x] Verify the intended Flora source repository, remote, and `main` branch
  before installing or running project code.
- [ ] Install only the runtime/build dependencies justified by the verified
  codebase and operating system.
- [ ] Define where application code, virtual environments/toolchains, logs,
  and scheduled research jobs may live; keep them separate from raw data.
- [ ] Establish basic resource limits and disk-space safeguards for batch jobs.

## 6. Final verification and handoff

- [ ] Re-run the sanitized baseline assessment and compare it with the initial
  state.
- [ ] Verify SSH access, firewall behavior, updates, time sync, and required
  services after a reboot.
- [ ] Verify `/var/data/flora` permissions and raw-data protection behavior.
- [ ] Record final decisions, remaining risks, recovery steps, and deferred
  work in `DECISIONS.md`.

## Deferred explicitly

- Live trading and broker execution
- Public-facing services and dashboards
- Kubernetes or other orchestration platforms
- Session semantics and event-motif research implementation
