#!/bin/bash
set -e
set -o pipefail

CONTEXT="${1}"
CLUSTER="${2}"

source ./hack/common.sh

# We expect a x.x.x.x/16 on the docker network bridge,
# and we need to have a /26 of that and assign it to metallb.
# Because we need non-overlapping ip ranges on metallb we use the cluster number (1, 2, 3)
# to calculate the last /26 subnets of that /16 range.
# cluster=1 gets the last /26
# cluster=2 gets the second to last /26
# etc. etc.
export METALLB_HOST_MIN=$(metallb_host_min "$CLUSTER")
export METALLB_HOST_MAX=$(metallb_host_max "$CLUSTER")
echo "using metallb range $METALLB_HOST_MIN-$METALLB_HOST_MAX"

kubectl apply --wait=true -f ./hack/env/metallb-native.yaml --context "${CONTEXT}"
kubectl wait deployment -n metallb-system controller --for condition=Available=True --timeout=90s --context "${CONTEXT}"
envsubst < ./hack/env/metallb.yaml | kubectl apply --wait=true -f - --context "${CONTEXT}"
