# Plan Patches

Plan patches can be used to customize the IAAS
environment and bosh director that is created by
`bbl up`.

In order to do so, you can use do the following:

```
mkdir some-env && cd some-env
bbl plan --name some-env
cp -r /path/to/patch-dir/. .
bbl up
```

A patch is a directory with a set of files
organized in the same hierarchy as the bbl-state dir.

## bosh-lite-gcp

To create a bosh-lite environment on gcp,
you can use the steps above with the
`bosh-lite-gcp` patch provided here.

## iso-segs-aws

To create an iso-segs environment on aws, you can:

```
mkdir some-env && cd some-env
bbl plan --name some-env --lb-type cf --lb-cert /path/to/lb.crt --lb-key /path/to/lb.key
cp /path/to/patch-dir/cloud-config/iso-segs-ops.yml cloud-config/
TF_VAR_isolation_segments="1" bbl up
```
