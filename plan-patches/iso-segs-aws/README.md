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
