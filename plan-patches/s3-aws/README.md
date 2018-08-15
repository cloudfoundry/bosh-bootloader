The intent here was to create an s3 bucket that works not only with the use-s3-blobstore.yml operations file, but also with the bbr script enable-backup-restore-s3-versioned.yml.  As such, the tfvars assumes you want versioning.  Expects that you want to replicate your s3 buckets to some region other than the primary.  

The only manual step is you'll need to activate the user created to get the aws keys/secrets to talk to the created s3 buckets.  And add appropriate bucket names to deployment variables.
