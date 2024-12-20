#!/bin/bash
set -e
set -o pipefail

CONTEXT=$1
CLUSTER=$2

source ./hack/common.sh

export METALLB_ADDRESS_CIDR=$(metallb_address_cidr $CLUSTER)
echo "Using MetalLB Address ${METALLB_ADDRESS_CIDR}"

kubectl apply --wait=true -f ./hack/env/metallb-native.yaml --context "${CONTEXT}"
kubectl wait deployment -n metallb-system controller --for condition=Available=True --timeout=90s --context "${CONTEXT}"
envsubst < ./hack/env/metallb.yaml | kubectl apply --wait=true -f - --context "${CONTEXT}"
