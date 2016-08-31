#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE}")/.." && pwd -P)"

cd "${project_dir}"

GOARCH="${GOARCH:-$(go env GOARCH)}"
GOOS="${GOOS:-$(go env GOOS)}"

pkg/${GOOS}_${GOARCH}/slackstack "$@"