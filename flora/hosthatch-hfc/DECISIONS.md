# HFC decisions and change record

Record approved operational decisions and completed changes here. Keep secrets,
IP addresses, private keys, and host fingerprints out of this file.

| Date | Area | Decision or change | Verification / recovery note |
| --- | --- | --- | --- |
| 2026-08-22 | Platform | HFC is a KVM virtual machine, not bare metal. | Verified with `systemd-detect-virt`. |
| 2026-08-22 | Platform | HFC runs Debian GNU/Linux 13.6 (trixie), with 3 provisioned vCPUs and 12 GB RAM. | Read-only baseline inspection. |
| 2026-08-22 | Access | HFC enrolled in the existing Tailscale tailnet. | Private SSH verified over the Tailscale address; preserve a recovery path before changing root SSH. |
| 2026-08-22 | Firewall | UFW enabled with a default-deny inbound policy; Cockpit public access blocked. | Cockpit verified through Tailscale; no public Cockpit allow rule. Tailscale UDP transport may remain public by design. |
| 2026-08-22 | Accounts | Created local `nnurry` sudo account for Cockpit and privilege elevation. | SSH remains key-only; account has no SSH key at this stage. |
