# HostHatch HFC: Flora research box

This directory tracks the setup and hardening of Flora's single in-scope remote
research server: the HostHatch HFC instance.

## Role and limits

- Offline market-data research only.
- Flora's raw vendor data is immutable evidence; normalized outputs are Parquet.
- No live trading, broker execution, public services, dashboards, or Kubernetes.
- Administration is performed through SSH using an SSH key; password login is
  not part of the intended access model.
- Administrative access is private through Tailscale/MagicDNS. Do not expose
  Cockpit or SSH as public application services.

## How to use these notes

Follow [HARDENING.md](HARDENING.md) in order. Each action should be performed
manually, understood, verified, and recorded in [DECISIONS.md](DECISIONS.md).

Never commit server IP addresses, private keys, passwords, API tokens, host
fingerprints, or other secrets to this repository.

## Known hardware

- KVM virtual machine (verified with `systemd-detect-virt`)
- 3 provisioned vCPUs, backed by an AMD Ryzen 9 9950X host platform
- 12 GB RAM
- 150 GB Gen 5 NVMe storage
