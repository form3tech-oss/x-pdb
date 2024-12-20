# Configuring Disruption Probes

x-pdb allows workload owners to define a disruption probe endpoint.

This endpoint is used by x-pdb to make a decision whether or not a disruption is allowed. Without the probe endpoint, x-pdb considers only pod readiness as a indicator of pod healthiness.

With the probe endpoint configured, x-pdb will ask the probe endpoint whether or not the disruption is allowed.

The probe endpoint is just a gRPC server, see details below.

It might be helpful to probe internal state of some workloads like databases to verify wether an eviction can happen or not.
Database Raft groups might become unavailable if a given pod is disrupted. In these cases workload owners might want to
block disruptions to happen, even if all pods are ready.

Example use-cases:

- assess health of a database cluster as a whole before allowing the deletion of a single pod, as this disruption may add more pressure on the database.
- assess the state of raft group leaders in a cluster before allowing an eviction
- assess the replication lag of database clusters
- assess if any long-running queries/jobs or backup are running, where a deletion can cause a problem

```yaml
apiVersion: x-pdb.form3.tech/v1alpha1
kind: XPodDisruptionBudget
metadata:
  name: opensearch
  namespace: opensearch
spec:
  minAvailable: 80%
  selector:
    matchLabels:
      k8s-app: opensearch
  probe:
    endpoint: opensearch-disruption-probe.opensearch.svc.cluster.local:8080
```

## TLS and authentication

At this point, the communication between x-pdb and the probe server does not use mutual TLS and only validates the server certificate presented by the probe endpoint and verifies it has been issued by the CA defined in `--controller-certs-dir`.

#### DisruptionProbe Server

The DisruptionProbe server allows the client to ask if a disruption for a given pod and XPDB resource is allowed.

```proto
// The DisruptionProbe service definition.
service DisruptionProbeService {
  // Sends a IsDisruptionAllowed request which will check if a given Pod
  // can be disrupted according to some specific rules.
  rpc IsDisruptionAllowed(IsDisruptionAllowedRequest) returns (IsDisruptionAllowedResponse) {}
}

// IsDisruptionAllowedRequest has the information to request a check for disruption.
message IsDisruptionAllowedRequest {
  // The name of the pod that is being disrupted.
  string pod_name = 1;

  // The namespaces of the pod that is being disrupted.
  string pod_namespace = 2;

  // The name of the XPodDisruptionBudget resource that was protecting the pod.
  string xpdb_name = 3;

  // The namespace of the XPodDisruptionBudget resource that was protecting the pod.
  string xpdb_namespace = 4;
}

// IsDisruptionAllowedRespobse has the information on wether a disruption is allowed or not.
message IsDisruptionAllowedResponse {
  // Information on wether disruption is allowed.
  bool is_allowed = 1;

  // Error information on why a disruption is not allowed.
  string error = 2;
}
```

You can check for a sample implementation on `cmd/testdisruptionprobe/main.go`.
