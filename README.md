# X-PDB

X-PDB allows you to define multi-cluster PodDisruptionBudgets and blocks evictions or Pod deletions when they are too disruptive.

This allows you to operate stateful workloads spanning multiple clusters and limit disruptions to a necessary minimum.
This is needed, because each cluster acts individually (evictions, rolling out updates etc) which may cause service disruptions simultaneously across clusters.

## 📙Documentation

X-PDB installation and reference documents are available at https://form3tech-oss.github.io/x-pdb.

👉 [Overview](https://form3tech-oss.github.io/x-pdb)

👉 [Getting Started](https://form3tech-oss.github.io/x-pdb/getting-started)

👉 [Configure X-PDB Resources](https://form3tech-oss.github.io/x-pdb/configuring-xpdb)

## Contributing

👉 [Developer Guide](https://form3tech-oss.github.io/x-pdb/developer-guide)

👉 [Code of Conduct](https://form3tech-oss.github.io/x-pdb/code-of-conduct)

## License

X-PDB is licensed under the [Apache License 2.0](./LICENSE).
