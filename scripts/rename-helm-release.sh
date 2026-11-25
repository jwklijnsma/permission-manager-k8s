#!/bin/sh

# Copyright (c) 2017-present SIGHUP s.r.l All rights reserved.
# Use of this source code is governed by a BSD-style
# license that can be found in the LICENSE file.

# Renames the GitHub release created for the current Helm chart version
# (tag "helm-chart-v<version>") to "Helm Chart v<version>".
#
# Usage: rename-helm-release.sh <github-token> [owner/repo]
#   The token can also be passed with the GITHUB_TOKEN / CR_TOKEN environment variables.
#   The repository defaults to $GITHUB_REPOSITORY (set by GitHub Actions) or sighupio/permission-manager.

set -eu

usage() {
    echo "Usage: $0 <github-token> [owner/repo]" >&2
    exit 2
}

for tool in curl jq yq; do
    if ! command -v "${tool}" >/dev/null 2>&1; then
        echo "error: required tool '${tool}' not found in PATH" >&2
        exit 1
    fi
done

CR_TOKEN="${1:-${CR_TOKEN:-${GITHUB_TOKEN:-}}}"
REPOSITORY="${2:-${GITHUB_REPOSITORY:-sighupio/permission-manager}}"

[ -n "${CR_TOKEN}" ] || usage

SCRIPT_DIR=$(CDPATH='' cd -- "$(dirname -- "$0")" && pwd)
CHART_FILE="${CHART_FILE:-${SCRIPT_DIR}/../helm_chart/Chart.yaml}"

if [ ! -f "${CHART_FILE}" ]; then
    echo "error: chart file '${CHART_FILE}' not found" >&2
    exit 1
fi

HELM_TAG=$(yq e '.version' "${CHART_FILE}")

if [ -z "${HELM_TAG}" ] || [ "${HELM_TAG}" = "null" ]; then
    echo "error: unable to read .version from '${CHART_FILE}'" >&2
    exit 1
fi

API="https://api.github.com/repos/${REPOSITORY}"

github() {
    curl --silent --show-error --fail --location \
        -H "Accept: application/vnd.github+json" \
        -H "Authorization: Bearer ${CR_TOKEN}" \
        -H "X-GitHub-Api-Version: 2022-11-28" \
        "$@"
}

if ! RELEASE_INFO=$(github "${API}/releases/tags/helm-chart-v${HELM_TAG}"); then
    echo "error: release for tag 'helm-chart-v${HELM_TAG}' not found in ${REPOSITORY}" >&2
    exit 1
fi

RELEASE_ID=$(printf '%s' "${RELEASE_INFO}" | jq -r '.id // empty')

if [ -z "${RELEASE_ID}" ]; then
    echo "error: unable to read the release id for tag 'helm-chart-v${HELM_TAG}'" >&2
    exit 1
fi

RELEASE_NAME="Helm Chart v${HELM_TAG}"
CURRENT_NAME=$(printf '%s' "${RELEASE_INFO}" | jq -r '.name // empty')

if [ "${CURRENT_NAME}" = "${RELEASE_NAME}" ]; then
    echo "release ${RELEASE_ID} is already named '${RELEASE_NAME}', nothing to do"
    exit 0
fi

github -X PATCH "${API}/releases/${RELEASE_ID}" \
    -d "$(jq -n --arg name "${RELEASE_NAME}" '{name: $name}')" >/dev/null

echo "release ${RELEASE_ID} renamed to '${RELEASE_NAME}'"
