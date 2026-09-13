# HFC K3s rollout

This is a learning-first plan for the first K3s installation on the HFC host.
It is intentionally a single-node cluster: HFC is both control plane and
schedulable worker. No Flora workload is deployed by this plan.

## Why K3s here

K3s includes its own container runtime (`containerd`), so HFC does not need a
separate Docker or Podman installation. It provides the Kubernetes API,
scheduler, kubelet, CNI, and runtime as a deliberately small single-node
installation.

## Chosen initial posture

- One K3s server, initialized with embedded etcd (`cluster-init: true`). This
  keeps a future multi-server path open without pretending the host is HA.
- HFC remains schedulable as a worker; no control-plane taint is added.
- The Kubernetes API advertises HFC's stable Tailscale address, not its public
  address.
- API TLS includes the Tailscale IP and MagicDNS hostname as SANs, so remote
  `kubectl` can verify the server certificate over Tailscale.
- Kubernetes Secrets are encrypted at rest in the embedded datastore.
- Traefik and ServiceLB are disabled: there is no public ingress or
  LoadBalancer-service use case.
- The generic K3s local-path provisioner is disabled. Flora will later use an
  explicit mount policy for `/var/data/flora`; `emptyDir` remains available.
- CoreDNS and metrics-server remain enabled.
- The kubeconfig remains root-readable only. Do not make the cluster-admin
  kubeconfig world-readable merely for convenience.

## Configuration to prepare

K3s loads `/etc/rancher/k3s/config.yaml` on every server start. The initial
file will be reviewed with the actual Tailscale values before creation:

```yaml
cluster-init: true
node-ip: <HFC_TAILSCALE_IP>
advertise-address: <HFC_TAILSCALE_IP>
tls-san:
  - <HFC_TAILSCALE_IP>
  - <HFC_MAGICDNS_NAME>
secrets-encryption: true
disable:
  - traefik
  - servicelb
  - local-storage
node-label:
  - flora.terabiome.io/role=hfc
  - flora.terabiome.io/tier=fast
```

Do not commit the actual tailnet hostname or Tailscale IP to this repository.

## Learning sequence

1. **Preflight:** inspect the stable Tailscale IP/MagicDNS name and confirm
   UFW blocks public TCP 6443 while permitting tailnet traffic.
2. **Configuration:** create and inspect the configuration file. Learn how
   server configuration differs from workload manifests.
3. **Installation:** install a current stable K3s release and allow systemd to
   start the `k3s` service.
4. **Local verification:** inspect the service, node readiness, system pods,
   control-plane datastore, and disk/RAM footprint.
5. **Remote verification:** use a Tailscale-only kubeconfig path from the
   administrator laptop; do not expose the API publicly.
6. **Policy verification:** confirm Traefik, ServiceLB, and generic local-path
   storage were not deployed.
7. **No-workload pause:** stop here until Flora's publication manifest and
   raw-namespace validation are complete.

## What this does not do

- It does not deploy a dashboard, public ingress, broker/trading component, or
  Flora workload.
- It does not mount `/var/data/flora` into any pod.
- It does not add offline fleet nodes or claim high availability.
- It does not replace HFC's host firewall or Tailscale policy.

## References

- [K3s server configuration](https://docs.k3s.io/cli/server)
- [K3s configuration file](https://docs.k3s.io/installation/configuration)
