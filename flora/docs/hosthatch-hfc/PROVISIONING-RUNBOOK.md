# HostHatch HFC provisioning runbook

Use this document to reproduce the **reviewed, manual** setup of a new Flora
HFC instance. Run commands only after adapting them to the new host and
verifying the stated preconditions.

Do not put IP addresses, login URLs, public keys, passwords, API tokens, or
Tailscale auth keys in this file.

## Operating rules

- Explain every command's purpose and effect before running it.
- Keep an existing root SSH session open while changing remote access or the
  firewall.
- Verify a new access path before removing an old one.
- Record deviations and post-change checks in `DECISIONS.md`.
- Treat existing local Flora data as rebuildable working state. Do not delete
  or clean it until the R2 retrieval/rebuild path is verified.

## HFC-001: Baseline assessment

**Purpose:** establish OS, virtualization, time, access, compute, storage, and
network exposure before configuration changes.

**Commands:**

```bash
whoami
hostnamectl
systemd-detect-virt
cat /etc/os-release
uname -r
uptime
timedatectl
w
nproc
lscpu
free -h
swapon --show
df -hT
findmnt -o TARGET,SOURCE,FSTYPE,OPTIONS
lsblk -o NAME,SIZE,TYPE,FSTYPE,MOUNTPOINTS,MODEL
ss -lntup
systemctl --type=service --state=running --no-pager
iptables -S
ip6tables -S
```

**Effect:** read-only inspection.

**Current HFC findings:** Debian 13.6, KVM VM, 3 vCPUs, 12 GB RAM, 150 GB
NVMe class storage, UTC/NTP synchronized, no swap, and existing completed
Flora output occupying roughly 60 GB under `/var/data/flora`. Do not assume a
new instance has the same capacity or existing data.

HFC local data is rebuildable working state. Raw processing inputs are pulled
from the authoritative R2 source; do not introduce a second HFC-data backup
pipeline unless that doctrine changes.

## HFC-002: Apply reviewed OS updates

**Purpose:** patch the base operating system without changing workload scope.

**Review first:**

```bash
apt list --upgradable
apt-get -s upgrade
```

**Apply after review:**

```bash
apt-get upgrade
```

**Current HFC result:** applied four Python 3.13 stable-update packages. No
kernel update or reboot was indicated.

## HFC-003: Install and enroll Tailscale

**Purpose:** establish a private, encrypted administration path before
restricting public application access.

**Preconditions:** an existing trusted tailnet; an owner/admin identity able to
approve the new device; direct root SSH remains open.

**Commands:**

```bash
install -d -m 0755 /usr/share/keyrings
curl -fsSL https://pkgs.tailscale.com/stable/debian/trixie.noarmor.gpg \
  | tee /usr/share/keyrings/tailscale-archive-keyring.gpg >/dev/null
curl -fsSL https://pkgs.tailscale.com/stable/debian/trixie.tailscale-keyring.list \
  | tee /etc/apt/sources.list.d/tailscale.list
apt-get update
apt-get install tailscale
tailscale up --accept-dns=false
tailscale ip -4
```

**Authentication:** open the login URL produced by `tailscale up` and approve
the host in the existing tailnet. Do not record that URL.

**Verify:** from a second local terminal, SSH to the displayed Tailscale IP
using the existing root key. Keep direct SSH open during the test.

**Current HFC result:** private root SSH verified over Tailscale. Tailscale SSH
was not enabled; standard OpenSSH remains the server.

## HFC-004: Create local Cockpit/sudo administrator

**Purpose:** provide a non-root local administration identity without adding a
new SSH credential yet.

**Commands:**

```bash
adduser --disabled-password --gecos "" nnurry
usermod -aG sudo nnurry
passwd nnurry
```

**Effect:** creates `nnurry`, grants sudo through Debian's `sudo` group, and
sets a password for Cockpit, sudo, and console use. It does not enable SSH
password login; SSH password authentication remains disabled.

**Verify:** sign in to Cockpit as `nnurry` and confirm sudo capability.

## HFC-005: Install Cockpit and restrict public ingress

**Purpose:** provide browser administration only through Tailscale, while
keeping a recovery-safe transition path.

**Install Cockpit:**

```bash
apt install cockpit
```

**Precondition:** verify that Cockpit is listening before firewalling:

```bash
ss -lntup | grep -E '(:22|:9090)'
```

Cockpit normally binds port 9090 on all interfaces. A host firewall must block
public access; do not rely on its login page as a network boundary.

**Install UFW and stage rules:**

```bash
apt install ufw
ufw default deny incoming
ufw default allow outgoing
ufw allow 41641/udp
ufw allow in on tailscale0
```

During a transition that retains public SSH, add this rule before enabling UFW:

```bash
ufw allow 22/tcp
```

Then enable enforcement:

```bash
ufw enable
```

After private SSH and provider recovery access are confirmed, remove the
temporary public SSH rule:

```bash
ufw delete allow 22/tcp
```

**Verify:**

```bash
ufw status verbose
```

Confirm Cockpit works at `https://<tailscale-ip>:9090` (or MagicDNS) and that
public port 9090 cannot be reached. Test public SSH closure only after a
working Tailscale SSH session and provider recovery route are retained.

**Current HFC result:** Cockpit works through Tailscale; public application
access is blocked. The final UFW status should be captured before copying this
state to another host.

## HFC-006: Verify the Flora source checkout

**Purpose:** ensure the host is using the intended Flora repository and branch
without disturbing research artifacts.

**Commands:**

```bash
git -C /root/Workspace/Code/flora status --short --branch
git -C /root/Workspace/Code/flora remote -v
```

**Effect:** read-only inspection. Do not run `git clean`, `git reset`,
`git pull`, or a fetch merely for this check.

**Expected:** `main` tracks `origin/main`; `origin` is
`github-nnurry:terabiome/flora`. Existing untracked notes, outputs, scripts, or
temporary artifacts must be reviewed—not deleted—before any Git cleanup.

**Current HFC result:** operator confirmed the checkout is current on `main`.

## Deferred hardening

- Record and test HostHatch console/recovery access.
- Add and test non-root SSH-key access before disabling root SSH login.
- Review Tailscale tailnet access policy before adding untrusted/shared devices.
- Verify R2 retention/availability and test the HFC input retrieval/rebuild
  path.
- Verify permissions, capacity, and local-retention behavior for the existing
  `/var/data/flora` layout.
