# Upgrade

In order to use a later version of `bbl` against an older
`bbl` environment/state directory, you will need to run
`bbl plan` before running `bbl up`.

`bbl plan` just writes the latest files and state directory structure.

`bbl up` is the applier. It will run `terraform apply`,
`bosh create-env`, and `bosh update-cloud-config`.

### Example

```
bbl5 up --lb-type cf --lb-cert cert --lb-key key --lb-domain domain.com

# some time passes

bbl6 plan --lb-type cf --lb-cert cert --lb-key key --lb-domain domain.com
bbl6 up

bbl6 destroy
```
