#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE}")/.." && pwd -P)"

cd "${project_dir}"

curl --fail --location --silent --show-error https://raw.githubusercontent.com/karlkfi/govendor/master/govendor.sh | bash -s -- "$@"
