# s3-blobstore-aws

This plan-patch adds terraform templates to create resources required to use
S3 for a CF blobstore.

## Steps:

1. Run `bbl plan`
1. Copy the contents of the terraform directory to the terraform sub-directory
   in your environment's state directory.
1. Run `bbl up`

## Permissions

To use this plan-patch, you will need to update the IAM policy for your bbl
IAM user to allow full access to S3 resources by adding the `s3:*` action.
