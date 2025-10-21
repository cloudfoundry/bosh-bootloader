#!/usr/bin/env bash

set -eu

fly -t bosh set-pipeline -p bosh-bootloader \
    -c ci/pipelines/bosh-bootloader.yml
