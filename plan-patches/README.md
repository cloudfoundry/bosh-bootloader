# Plan Patches

Plan patches can be used to customize the IAAS
environment and bosh director that is created by
`bbl up`.

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

## iso-segs-aws

To create an isolation segment on aws, the files in this directory
should be copied to your bbl state directory.

The steps might look like such:

```
mkdir banana-env && cd banana-env

bbl plan --name banana-env --lb-type cf --lb-cert lb.crt --lb-key lb.key

cp -r bosh-bootloader/plan-patches/iso-segs-aws/. .

TF_VAR_isolation_segments="1" bbl up
```


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

## tf-backend-gcp

Stores the terraform state in a bucket in Google Cloud Storage.

```
cp -r bosh-bootloader/plan-patches/tf-backend-gcp/. .
```

Since the backend configuration is loaded by Terraform extremely early (before
the core of Teraform can be initialized), there can be no interplations in the backend
configuration template. Instead of providing vars for the bucket to a `gcs_backend_override.tfvars`,
the values for the bucket name, credential path, and state prefix must be provided directly
in the backend configuration template.

Modify `terraform/gcs_backend_override.tf` to provide the name of the bucket, the path to
the GCP service account key, and a prefix for the environment inside the bucket.

Then you can bbl up.

```
bbl up
```

## iam-profile-aws

To use an existing iam instance profile on aws, the files in this directory
should be copied to your bbl state directory.

The steps might look like such:

```
mkdir banana-env && cd banana-env

bbl plan --name banana-env

cp -r bosh-bootloader/plan-patches/iam-profile-aws/. .
```

Write the name of the iam instance profile in `vars/iam.tfvars`.

```
bbl up
```

Providing the iam instance profile the bosh director means that the iam policy for
the user you give to bbl does not require `iam:*` permissions.
