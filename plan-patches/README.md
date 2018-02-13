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

## restricted-instance-groups-gcp

To create two instance groups instead of an instance group for every zone on gcp,
you can use the steps above with the `restricted-instance-groups-gcp` patch
provided here.

## iso-segs-gcp

Creates a single routing isolation segment on GCP, including dedicated load balancers and firewall rules.

```
cp -r bosh-bootloader/plan-patches/iso-segs-gcp/. some-env/
bbl up
```

Disclaimer: this is a testing/development quality patch.  It has not been subject to a security review -- the firewall rules may not be fully locked down.
Please don't run it in production!

