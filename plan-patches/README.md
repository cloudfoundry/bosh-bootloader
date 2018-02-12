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


## iso-segs-aws

To create an iso-segs environment on aws, you can:

```
mkdir some-env && cd some-env
bbl plan --name some-env --lb-type cf --lb-cert /path/to/lb.crt --lb-key /path/to/lb.key
cp /path/to/patch-dir/cloud-config/iso-segs-ops.yml cloud-config/
TF_VAR_isolation_segments="1" bbl up
```

## iam-profile-aws

To use an existing iam instance profile for the bosh director on aws, you can:

```
mkdir some-env && cd some-env
bbl plan --name some-env
cp -r bosh-bootloader/plan-patches/iam-profile-aws/. some-env/

# write the name of the iam instance profile in the vars/iam.tfvars file

bbl up
```

Providing the iam instance profile the bosh director means that the iam policy for
the user you give to bbl does not require `iam:*` permissions.
