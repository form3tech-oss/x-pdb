# X-PDB

X-PDB allows you to define multi-cluster PodDisruptionBudgets and blocks evictions or Pod deletions when they are too disruptive.

This allows you to operate stateful workloads spanning multiple clusters and limit disruptions to a necessary minimum.
This is needed, because each cluster acts individually (evictions, rolling out updates etc) which may cause service disruptions simultaneously across clusters.

## Eviction Process

X-PDB hooks into the DELETE `pods` or CREATE `pods/eviction` API calls using a `ValidatingWebhookConfiguration`.
It will acquire locks on all clusters to prevent race conditions and read the state (expected pod count & healthy pod count) from all clusters and compute if a eviction/deletion is allowed.

X-PDB need to talk to the other X-PDB pods the other clusters to read the remote state.

### Locking mechanism

X-PDB acquires a lock on the remote clusters using a HTTP API.
The lock is valid for a specific `namespace/selector` combination and it has a `leaseHolderIdentity`. This is the owner of the given lock.

The lock is **valid for 5 seconds**. After that it can be re-acquired or taken over by a different holder.
The lock prevents a race condition which occur if multiple evictions happen simultaneously across clusters which would lead to inconsistent data and wrong decisions. E.g. a read can happen while a eviction is being processed which would lead to multiple evictions happen at the same time that could break the pod disruption budget.

We leave the lock as it is and DO NOT unlock it after the admission webhook have finished processing.
Once it expires it can be re-acquired or taken over. We rely on the caller to retry the eviction or deletion of a Pod.


**Why 5 seconds? - Why leave it locked when we've finished processing?**

We need to lock evictions for a period of time **after** we have returned the admission result to allow kube-apiserver to update the Pod (set the `deletionTimestamp`).

The lease duration should be higher than the sum of the following duration:
  - the round-trip latency across all clusters (1-2 seconds)
  - the processing time of the x-pdb http server (<1 second)
  - the time kube-apiserver needs to process the x-pdb admission control response
    and the time it takes until the desired action (evict/delete pod) is observable through the kube-apiserver (1-2 seconds)
  - a generous surcharge (1-... seconds)

### State server

In order for x-pdb servers to communicate between each other they expose a [gRPC server](./protos/state/state.proto) which has the following API:

```proto
service State {

  rpc Lock(LockRequest) returns (LockResponse) {}
  rpc Unlock(UnlockRequest) returns (UnlockResponse) {}
  rpc GetState(GetStateRequest) returns (GetStateResponse) {}
}
```

## Threat Modelling

- [Threat Modelling inventory](/threat-modelling/TMinventory.md)

## Observability & monitoring

## Development
### Running tests

Simple tests can be ran with `make test`.

Integration tests can be ran with the following commands.
This will create three `kind` clusters which are connected through `metallb`.

```
make multi-cluster
make e2e
```

The e2e tests deploy a test pod which allows us to control their readiness probe, see `cmd/testapp` for details.

When developing `x-pdb` e2e tests you need to re-build and deploy the Pod using `make deploy`.

