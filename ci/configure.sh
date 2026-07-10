#!/usr/bin/env bash

set -eu

pipeline="$(mktemp)"
trap 'rm -f "${pipeline}"' EXIT

ytt -f ci/pipelines/bosh-bootloader.yml --ignore-unknown-comments > "${pipeline}"

fly -t bosh set-pipeline -p bosh-bootloader \
    -c "${pipeline}"
