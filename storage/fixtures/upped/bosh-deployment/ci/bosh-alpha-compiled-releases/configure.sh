#!/usr/bin/env bash

set -eu

fly -t production set-pipeline \
 -p bosh-alpha-compiled-releases \
 -c ./pipeline.yml