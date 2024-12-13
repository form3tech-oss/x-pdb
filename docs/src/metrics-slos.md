# Metrics & SLOs

x-pdb exposes Prometheus metrics on the `/metrics` path. To enable it, set the `serviceMonitor.enabled` Helm flag to true.

## Metrics
### x-pdb

| Name                                       | Type  | Description                                                |
|--------------------------------------------|-------|------------------------------------------------------------|
| `pod_eviction_rejected`   | Counter | Represents the number of eviction which have been rejected through x-pdb. |
| `pod_matches_multiple_xpdbs` | Counter | A eviction attempt for a pod has been observed which matches multiple XPDBs. This is a invalid configuration and must be fixed. |
| `lock_errors` | Counter | Counter that represents the number of errors when obtaining locks for xpdb.|

### grpc metrics

x-pdb exposes GRPC metrics for both [client](https://github.com/grpc-ecosystem/go-grpc-middleware/blob/ba6f8b95444c087a9ed0af5b78a5e56cad57964b/providers/prometheus/client_metrics.go#L35-L57) and [server](https://github.com/grpc-ecosystem/go-grpc-middleware/blob/ba6f8b95444c087a9ed0af5b78a5e56cad57964b/providers/prometheus/server_metrics.go#L30-L52) which allow you to get insights into latency and availability of the remote x-pdb servers.

## SLOs

### Availability

There should be at least one pod ready to serve traffic at any time, preferably measured from both `kube-apiserver` and `x-pdb` on other clusters.

```
sum(increase(apiserver_admission_webhook_fail_open_count{name=~".*x-pdb.*"}[5m]))
```

### Latency

The amount of time x-pdb needs to respond to a admission webhook, preferably measured from the kube-apiserver. It should take less than 150ms for `x-pdb` to respond to admission requests on the p99. The threshold may vary in your environment, depending on the cross-cluster latency.

```
histogram_quantile(0.99,
    sum(rate(apiserver_admission_webhook_admission_duration_seconds_bucket{name=~".*x-pdb.*"}[5m])) by (le, name)
)
```
