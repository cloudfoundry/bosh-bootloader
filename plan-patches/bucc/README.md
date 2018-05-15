# Patch: [BUCC](https://github.com/starkandwayne/bucc)

Steps to deploy BUCC with bbl:

## Initialize a new bbl environment repo

The required patch files are maintained in the BUCC repo, and need to be symlinked into the envrionment repo.
To keep track of the BUCC version, we will be using a git submodule.

```
export BBL_IAAS=aws|gcp|azure
export BBL_ENV_NAME=banana-env
mkdir $BBL_ENV_NAME && cd $BBL_ENV_NAME && git init
bbl plan -lb-type concourse
git submodule add https://github.com/starkandwayne/bucc.git bucc
ln -s bucc/bbl/*-director-override.sh .
ln -sr bucc/bbl/terraform/$BBL_IAAS/* terraform/
bbl up
eval "$(bbl print-env)"
eval "$(bucc/bin/bucc env)"
```

## Update BUCC

The master branch of BUCC is always the latest stable release.
Make sure to check the [releases notes](https://github.com/starkandwayne/bucc/releases) for additional instructions.

```
git submodule update --remote bucc
bbl up
```

## Usage

Just use bucc [as normal](https://github.com/starkandwayne/bucc#using-bucc).
