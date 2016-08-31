#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE}")/.." && pwd -P)"

cd "${project_dir}"

# Must be run from the go project dir TODO: validate?
# Project dir must be mountable into docker (under $HOME on Mac) TODO: validate?

# Determine the project import path from the root ImportComment
# Optionally specify via GO_PROJECT_IMPORT
if [ -z "${PROJECT_IMPORT:-}" ]; then
  PROJECT_IMPORT="$(
    docker run --rm \
        -v "$(pwd):/project" \
        -w '/project' \
        golang:1.6.2-alpine \
        go list -f '{{.ImportComment}}'
  )"
fi

# Optionally pass in SSH key to pull dependencies from private repos
PRIVATE_KEY_PARAM=""
if [ -n "${PRIVATE_KEY:-}" ]; then
  PRIVATE_KEY_PARAM="-v ${PRIVATE_KEY:-}:/root/.ssh/id_rsa"
fi

docker run --rm -i \
       $(tty -s && echo '-t') \
       -v "$(pwd):/go/src/${PROJECT_IMPORT}" \
       ${PRIVATE_KEY_PARAM} \
       -w "/go/src/${PROJECT_IMPORT}" \
       karlkfi/govendor "$@"
