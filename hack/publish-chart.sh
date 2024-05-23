#!/usr/bin/env bash

set -o errexit
set -o pipefail
set -o nounset
if [[ "${TRACE-0}" == "1" ]]; then
    set -o xtrace
fi

GITHUB_REPOSITORY_OWNER=${GITHUB_REPOSITORY_OWNER-ionos-cloud}
OCI_HELM_CHART_REPO=oci://ghcr.io/${GITHUB_REPOSITORY_OWNER}/helm-charts

CHART_NAME=ionoscloud-blockstorage-csi-driver
CHART_VERSION=${CHART_VERSION-$(yq -r '.version' "charts/${CHART_NAME}/Chart.yaml")}
CHART_PACKAGE=${CHART_NAME}-${CHART_VERSION}.tgz

if ! helm show chart "${OCI_HELM_CHART_REPO}/${CHART_NAME}" --version "${CHART_VERSION}" >/dev/null; then
    helm package "charts/${CHART_NAME}"
    if [[ -n ${DRY_RUN-} ]]; then
        echo "Pushed: ${OCI_HELM_CHART_REPO}/${CHART_NAME}:${CHART_VERSION} (dry-run)"
        exit
    fi
    helm push "${CHART_PACKAGE}" "${OCI_HELM_CHART_REPO}"
fi
