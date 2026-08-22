# Flora infrastructure

## Current setup scope

Set up and harden **one HostHatch HFC instance** as Flora's single research box.

The box is for offline market-data research only. It is not a live trading,
broker-execution, public-service, Kubernetes, or dashboard host.

## Known machines

### In scope: HostHatch HFC

- One remote HFC instance at HostHatch.
- Target role: Flora research box.
- KVM VM provisioned with 3 vCPUs, 12 GB RAM, and 150 GB Gen 5 NVMe storage;
  backed by an AMD Ryzen 9 9950X host platform.

### Offline lab machines (out of scope for this setup)

- All run openSUSE MicroOS, an immutable operating system, on bare metal.
- One Xeon box: 2 × Intel Xeon E5-2686 v4; 8 × 32 GB DDR4-2133 RAM.
- Four Lenovo Tiny PCs:
  - One M70q with a 10th-generation Intel Core i5 and 2 × 16 GB DDR4-2400 RAM.
  - Three M710q units with 8th-generation Intel Core i5 CPUs and 2 × 16 GB
    DDR4-2400 RAM each.

Future fleet configuration must respect MicroOS's transactional update and
reboot model; do not assume mutable-host package-management workflows.

## Flora doctrine

- Preserve raw vendor data as immutable evidence.
- Normalize data to Parquet.
- Build causal features; keep labels separate.
- Research concrete futures contracts, not stitched continuous ZC.
- Avoid lookahead bias and overfitting.

Current market-data focus: CBOT Corn futures (ZC), using Databento
`GLBX.MDP3` MBP-1 data.

Authoritative raw vendor data is retained in R2. HFC local Flora data is
rebuildable working state: the host pulls raw inputs from R2 before processing.
