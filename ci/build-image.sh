#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE}")/.." && pwd -P)"

cd "${project_dir}"

docker build -t karlkfi/slackstack .
