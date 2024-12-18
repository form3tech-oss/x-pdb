# Developer Guide

## Getting Started

You must have a working [Go environment](https://golang.org/doc/install) and
then clone the repo:

```shell
git clone https://github.com/form3tech-oss/x-pdb.git
cd x-pdb
```

You'll need the following tools to work with the project:

- [Kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl)
- [Kind](https://kind.sigs.k8s.io/docs/user/quick-start/#installation)
- [Helm](https://helm.sh/docs/intro/install/)
- [Kubebuilder](https://book.kubebuilder.io/quick-start.html#installation)
- [Buf](https://buf.build/docs/installation/)

## Build and Test

The project uses the `make` build system. It'll run code generators, tests and static code analysis.

### Development environment

The development environment is based of 3 kind clusters that are connected together through some loadbalancers using MetalLB.

X-PDB on each cluster will expose a LoadBalancer service which the IP is going to be provisioned by MetalLB. That IP is going to be available to all the clusters, allowing X-PDB to talk with other X-PDB services in other clusters.

To spin up the development environment you should run the following command:

```bash
make multi-cluster
```

You should run the following command to install x-pdb on all the 3 kind clusters:

```bash
make deploy
```

After it is installed you are able to create workloads and x-pdb resources to test out the features.

### Testing

#### Unit tests

X-PDB unit tests can be run with the following command:

```bash
make test
```

#### E2E Tests

E2E test suite will need the development environment to be up and running.
You will need to build and deploy some testing applications by running the following command.

```bash
make deploy-e2e
```

After the testing applications are installed you can run the E2E test suite by running the following command:

```bash
make e2e
```

### Linting

Before commiting your changes ensure that codebase is linted.

```bash
make fmt
make proto-format
make lint
make helm-lint
make proto-lint
```

### Working with gPRC

The gPRC protobuf contracts are declared in:

- `proto/disruptionprobe/v1/disruptionprobe.proto`
- `proto/state/v1/state.proto`

These contracts are managed by [buf](https://buf.build/docs/cli/).
It will allow us to easily manage everything related with gRPC.

After you make changes to proto contracts please run the following commands to ensure all changes valid:

```bash
make proto-format
make proto-lint
make proto-breaking
make proto-generate
```

## Documentation

Documentation for this project is done by [mkdocs](https://www.mkdocs.org/).

To test out the changes while you're changing documentation run the following command:

```bash
cd docs
make live-docs
```
