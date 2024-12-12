#!/bin/bash
ROOT="$(git rev-parse --show-toplevel)"
pushd $ROOT 2> /dev/null || true

CLUSTER=$1
helm upgrade -i x-pdb ./charts/x-pdb -f $ROOT/hack/env/xpdb-$CLUSTER-values.yaml --kube-context=kind-x-pdb-$CLUSTER
kubectl rollout restart deploy/x-pdb --context=kind-x-pdb-$CLUSTER
