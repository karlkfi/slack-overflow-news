#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE}")/.." && pwd -P)"

cd "${project_dir}"

GOARCH="${GOARCH:-$(go env GOARCH)}"
GOOS="${GOOS:-$(go env GOOS)}"

# Build local first for fast feedback
GOOS="${GOOS}" GOARCH="${GOARCH}" bash -c "go build -o \"pkg/\${GOOS}_\${GOARCH}/slack-overflow-news\""

# Build server binary
if [[ "${GOOS}" != "linux" ]] || [[ "${GOARCH}" != "amd64" ]]; then
  GOOS=linux GOARCH=amd64 bash -c "go build -o \"pkg/\${GOOS}_\${GOARCH}/slack-overflow-news\""
fi
