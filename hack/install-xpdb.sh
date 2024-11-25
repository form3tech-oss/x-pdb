#!/bin/bash
set -e
set -o pipefail

CONTEXT="${1}"
CLUSTER="${2}"

source ./hack/common.sh

# every cluster needs a different set of endpoints:
# this_address is the address of the x-pdb server of a given cluster
# remote_endpoints are comma-separated values (URLs) which point to the _remote_ clusters
this_address=""
remote_endpoints=""
for i in 1 2 3; do
  METALLB_HOST_MIN=$(metallb_host_min $i)
  if [[ "$i" -eq "$CLUSTER" ]]; then
    this_address="$METALLB_HOST_MIN"
    continue
  fi

  if [[ ! -z "$remote_endpoints" ]]; then
    remote_endpoints="${remote_endpoints}\," # comma needs to be escaped for helm
  fi
  remote_endpoints="${remote_endpoints}$METALLB_HOST_MIN:443"
done

echo "=============================="
echo "remote endpoints: $remote_endpoints"
echo "this address: $this_address"
echo "=============================="

if [[ -z "$this_address" ]]; then
  echo "error: cluster $CLUSTER not found in range 1-3: could not find appropriate endpoint."
  exit 1
fi

helm upgrade -i x-pdb ./helm \
  -f "hack/env/xpdb-values.yaml" \
  --set controller.clusterID="${CLUSTER}" \
  --set controller.remoteEndpoints="$remote_endpoints" \
  --set controller.tls.certManager.ipAddresses="{$this_address}" \
  --set service.controller.loadBalancerIP="$this_address" \
  --kube-context="${CONTEXT}"

kubectl rollout restart deploy/x-pdb --context="${CONTEXT}"
kubectl wait deployment x-pdb --for condition=Available=True --timeout=90s --context="${CONTEXT}"
