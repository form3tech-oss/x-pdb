# Configuring XPodDisruption Resource

## Nomenclature

| term            | description                                                                         |
| --------------- | ----------------------------------------------------------------------------------- |
| local cluster   | This is the Kubernetes cluster where a pod deletion or eviction is being requested. |
| remote clusters | These are the clusters which are not handling the pod eviction / deletion.          |

## Configuration and Behavior

X-PDB works under the assumption that your workloads are structured in a similar way across clusters, i.e. that pods that you want to protect sit in the same namespace no matter which cluster you look at.

The `XPodDisruptionBudget` resources looks and feels just like `PodDisruptionBudget` resources:

- A label selector `.spec.selector` to specify the set of pods to which it applies. The `selector` applies to all clusters for decision making, not just the local one where pod eviction/deletion is happening.

- `.spec.minAvailable` which is a description of the number of pods from that set that must still be available after the eviction, even in the absence of the evicted pod. minAvailable can be either an absolute number or a percentage.
- `.spec.maxUnavailable` which is a description of the number of pods from that set that can be unavailable after the eviction. It can be either an absolute number or a percentage.

You can specify only one of `maxUnavailable` and `minAvailable` in a single PodDisruptionBudget. `maxUnavailable` can only be used to control the eviction of pods that have an associated controller managing them.

In addition to that, `XPodDisruptionBudget` has the following fields:

- `.spec.suspend` which allows you to disable the XPDB resource. This allows all pod deletions/evictions. It is intended to be used as a break-glass procedure to allow engineers to take manual action. The suspension is configured on a per-cluster basis and affects only local pods. I.e. other clusters that run x-pdb will not be able to evict pods if there isn't enough disruption budget available globally.
- `.spec.probe` that allows workload owners to define a [disruption probe](./configuring-disruption-probes.md) endpoint. Without a probe, x-pdb will only consider pod readiness as an indicator of healthiness and compute the disruption verdict based on that. With `.spec.probe`, x-pdb considers the response of the probe endpoint as well.

It is irrelevant for `x-pdb` if the remote cluster has a `XPodDisruptionBudget` resource and whether or not the configuration match.

The user is supposed to deploy the `XPodDisruptionBudget` to all clusters. It may lead to unexpected disruptions when the resource is missing.

```yaml
apiVersion: x-pdb.form3.tech/v1alpha1
kind: XPodDisruptionBudget
metadata:
  name: kube-dns
  namespace: kube-system
spec:
  # Specify either `minAvailable` or `maxUnavailable`
  # Both percentages and numbers are supported
  minAvailable: 80%
  selector:
    matchLabels:
      k8s-app: kube-dns
```

## gRPC State Server

In order for x-pdb servers to communicate between each other they expose a gRPC state server interface with the following APIs. It allows x-pdb to asses the health of pods on remote clusters.

The communication between x-pdb servers is secured using mutual TLS. The certificate directory can be configured with `--controller-certs-dir` which is supposed to contain `ca.crt`, `tls.crt` and `tls.key` files.

```proto
// State is the service that allows x-pdb servers to talk with
// each other.
service State {
  // Acquires a lock on the local cluster using the specified leaseHolderIdentity.
  rpc Lock(LockRequest) returns (LockResponse) {}

  // Frees a lock on the local cluster if the lease identity matches.
  rpc Unlock(UnlockRequest) returns (UnlockResponse) {}

  // Calculates the expected count based off the Deployment/StatefulSet/ReplicaSet number of replicas or - if implemented - a `scale` sub resource.
  rpc GetState(GetStateRequest) returns (GetStateResponse) {}
}
```
