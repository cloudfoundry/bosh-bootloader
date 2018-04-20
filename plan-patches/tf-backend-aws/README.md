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

