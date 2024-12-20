#!/bin/bash

function docker_bridge_base_cidr() {
  local DOCKER_CIDR=$(docker network inspect -f json kind | jq -r '.[0].IPAM.Config[] | select(.Subnet | contains(":") | not) | .Subnet')
  echo $DOCKER_CIDR | cut -d'.' -f1-2
}

function metallb_address_cidr() {
    local CLUSTER_ID=$1
    local DOCKER_BASE_CIDR=$(docker_bridge_base_cidr)
    echo "$DOCKER_BASE_CIDR.$((250+$CLUSTER_ID)).250/28"
}

function x_pdb_address() {
    local CLUSTER_ID=$1
    local DOCKER_BASE_CIDR=$(docker_bridge_base_cidr)
    echo "$DOCKER_BASE_CIDR.$((250+$CLUSTER_ID)).251"
}
