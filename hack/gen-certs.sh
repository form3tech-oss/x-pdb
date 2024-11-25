#!/bin/bash
set -e
set -o pipefail

ROOT="$(git rev-parse --show-toplevel)"
mkdir -p $ROOT/hack/certs 2> /dev/null
pushd $ROOT/hack/certs 2> /dev/null

rm tls.* 2> /dev/null || true
rm ca.* 2> /dev/null || true

# Generate ca.key
openssl genrsa -out ca.key 4096

# Generate ca.crt
openssl req -x509 -new -sha512 \
-key ca.key -days 365 \
-config ./../ca.conf \
-out ca.crt
