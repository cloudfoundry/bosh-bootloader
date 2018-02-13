# Patch: bosh-lite-gcp

Plan patches can be used to customize the IAAS
environment and bosh director that is created by
`bbl up`.

To create a bosh lite on gcp, the files in this directory
should be copied to your bbl state directory.

The steps might look like such:

```
mkdir banana-env && cd banana-env

bbl plan --name banana-env

cp -r bosh-bootloader/plan-patches/bosh-lite-gcp/. .

bbl up
```

