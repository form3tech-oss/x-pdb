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
  name: kube-dns
  namespace: kube-system
spec:
  minAvailable: 80%
  selector:
    matchLabels:
      k8s-app: kube-dns
  probe:
    endpoint: opensearch-disruption-probe.opensearch.svc.cluster.local:8080
```

## TLS and authentication

At this point, the communication between x-pdb and the probe server does not use mutual TLS and only validates the server certificate presented by the probe endpoint and verifies it has been issued by the CA defined in `--controller-certs-dir`.

#### DisruptionProbe Server

The DisruptionProbe server allows the client to ask if a disruption for a given pod and XPDB resource is allowed.

```proto
service DisruptionProbe {
  // Sends a IsDisruptionAllowed request
  rpc IsDisruptionAllowed(IsDisruptionAllowedRequest) returns (IsDisruptionAllowedResponse) {}
}

// The request message containing the user's name.
message IsDisruptionAllowedRequest {
  string pod_name = 1;
  string pod_namespace = 2;
  string xpdb_name = 3;
  string xpdb_namespace = 4;
}

// The response message containing the greetings
message IsDisruptionAllowedResponse {
  bool is_allowed = 1;
  string error = 2;
}

```

