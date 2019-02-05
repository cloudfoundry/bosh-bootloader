# gcs-blobstore-gcp

This plan-patch adds terraform templates to create resources required to use
GCS (Google Cloud Storage) for a CF blobstore.

## Steps:

1. Run `bbl plan`
1. Copy the contents of the terraform directory to the terraform sub-directory
   in your environment's state directory.
1. Run `bbl up`
