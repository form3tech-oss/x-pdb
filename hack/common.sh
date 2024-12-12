#!/bin/bash

function docker_bridge_cidr() {
  docker network inspect -f json kind | jq -r '.[0].IPAM.Config[] | select(.Subnet | contains(":") | not) | .Subnet'
}

function metallb_cidr() {
  CLUSTER_ID="${1}"
  NET_NUM=$((1024 - $CLUSTER_ID))
  BRIDGE_CIDR=$(docker_bridge_cidr "${CLUSTER_ID}")
  echo "cidrsubnet(\"$BRIDGE_CIDR\", 10, $NET_NUM)" | terraform console | sed 's/"//g'
}

function host_min() {
  CIDR="${1}"
  echo "cidrhost(\"$CIDR\", 1)" | terraform console | sed 's/"//g'
}

function host_max() {
  CIDR="${1}"
  echo "cidrhost(\"$CIDR\", -2)" | terraform console | sed 's/"//g'
}

function metallb_host_min() {
  CLUSTER_ID="${1}"
  METALLB_CIDR=$(metallb_cidr "${CLUSTER_ID}")
  host_min "$METALLB_CIDR"
}

function metallb_host_max() {
  CLUSTER_ID="${1}"
  METALLB_CIDR=$(metallb_cidr "${CLUSTER_ID}")
  host_max "$METALLB_CIDR"
}
