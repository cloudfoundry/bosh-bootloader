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

## tf-backend-aws

Stores the terraform state in a given bucket on Amazon S3.

```
cp -r bosh-bootloader/plan-patches/tf-backend-aws/. .
```

Since the backend configuration is loaded by Terraform extremely early (before
the core of Teraform can be initialized), there can be no interplations in the backend
configuration template. Instead of providing vars for the bucket to an `s3_backend_override.tfvars`,
the values for the bucket name, region, and key for the state must be provided directly
in the backend configuration template.

Modify `terraform/s3_backend_override.tf` to provide the name and region of the bucket,
as well as the key to write the terraform state to.

Then you can bbl up.

```
bbl up
```
