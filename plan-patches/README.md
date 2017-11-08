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

For instance to create a bosh-lite environment
on gcp, you can use the steps above with the
`bosh-lite-gcp` patch provided here.
