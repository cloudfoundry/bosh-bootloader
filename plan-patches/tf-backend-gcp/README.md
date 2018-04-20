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

