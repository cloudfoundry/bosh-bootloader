# Patch: 1-az-aws

This patch is for using a single availability zone from `BBL_AWS_REGION`.

Steps:

1. Run `bbl plan`.

1. Copy the `vars/zone.tfvars` into `${BBL_STATE_DIR}/vars/`.

1. Use the `bbl.tfvars` to see the list of possible availability zones. Choose one and set it in the array in `zone.tfvars`.

1. Run `bbl up`.

**Note:** If you know what availability zones are available in your chosen
AWS region, you can also copy this patch and choose an availability zone
to put in `zone.tfvars` and you can run `bbl up` without first running `bbl plan`.
