#!/usr/bin/env bash

set -o errexit
set -o nounset
set -o pipefail

project_dir="$(cd "$(dirname "${BASH_SOURCE}")/.." && pwd -P)"

cd "${project_dir}"

GOARCH="${GOARCH:-$(go env GOARCH)}"
GOOS="${GOOS:-$(go env GOOS)}"
CMD="pkg/${GOOS}_${GOARCH}/slack-overflow-news"

if [ -f ".env" ]; then
  env $(cat .env | xargs) ${CMD} "$@"
else
  ${CMD} "$@"
fi
