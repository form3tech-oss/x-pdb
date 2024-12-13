# X-PDB

X-PDB allows you to define multi-cluster PodDisruptionBudgets and blocks evictions or Pod deletions when they are too disruptive.

This allows you to operate stateful workloads spanning multiple clusters and limit disruptions to a necessary minimum.
This is needed, because each cluster acts individually (evictions, rolling out updates etc) which may cause service disruptions simultaneously across clusters.

Please refer to the documentation at https://form3tech-oss.github.com/x-pdb

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

