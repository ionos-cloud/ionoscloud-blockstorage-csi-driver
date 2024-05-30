#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
if [[ "${TRACE-0}" == "1" ]]; then
    set -o xtrace
fi

BASE_DIR=$(mktemp -d)
trap 'rm -rf $BASE_DIR' EXIT

if [[ $# -eq 0 ]]; then
    TAG=$(yq -r '.snapshotter.image.tag' charts/ionoscloud-csi-driver/values.yaml)
else
    TAG=$1
fi

git clone https://github.com/kubernetes-csi/external-snapshotter --no-checkout "$BASE_DIR"
(
    cd "$BASE_DIR"
    git sparse-checkout set client/config/crd
    git checkout "$TAG"
)
cp "$BASE_DIR/client/config/crd/"snapshot.*.yaml charts/ionoscloud-csi-driver/crds/
