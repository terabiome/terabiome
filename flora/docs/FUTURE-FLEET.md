# Proposed Flora fleet topology

This is a future-state design note. It does not authorize or schedule a
Kubernetes/K3s installation.

## Intended roles

```text
HostHatch HFC (always on)
├── K3s control plane
├── Worker node
├── Latency-sensitive / interactive Flora workloads
└── Batch work when capacity is available

Offline bare-metal MicroOS machines (powered on as needed)
└── Worker nodes for latency-insensitive batch workloads
```

The HFC's compute capacity is not reserved exclusively for latency-sensitive
work. It can run batch jobs normally. When a latency-sensitive workload is to
run, the offline workers are powered on and batch work should preferentially
run there, preserving HFC capacity and responsiveness.

## Scheduling policy to design later

- Label the HFC for both `fast` and `batch` eligibility.
- Label offline nodes for `batch` eligibility.
- Target latency-sensitive workloads explicitly to the HFC and give them a
  higher Kubernetes PriorityClass.
- Prefer offline workers for batch Jobs whenever they are available.
- When HFC capacity must be protected, prevent new batch placement there and
  move or allow preemption of interruptible batch work.

Kubernetes priority affects scheduling and preemption, but does not by itself
guarantee that batch Jobs choose offline machines. The final design must choose
an explicit mechanism for latency-sensitive periods, such as temporarily
cordoning HFC for new batch work, changing batch Job affinity at submission, or
using an admission/queue layer. The mechanism should be simple, observable,
and reversible.

## Data and failure constraints

- Batch workers are intermittent. Jobs must be idempotent, retry-safe, and use
  durable inputs and outputs.
- Keep raw vendor data immutable. Do not rely on a container writable layer for
  research evidence.
- Keep data movement intentional: HFC has 150 GB NVMe, while the Xeon has the
  8 TB HDD archive. Avoid needless transfers of large raw datasets through HFC.
- The control plane is initially a single point of failure by design. Back up
  its state and maintain a tested rebuild procedure before depending on it.
