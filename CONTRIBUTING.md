# Contributing to BOSH-Bootloader

The Infrastructure team accepts contributions via pull request against master.

To verify your changes before submitting a pull request:

1. Run unit tests

  ```
    ~/bosh-bootloader/scripts/test
  ```

2. Run acceptance tests

  ```
    ~/bosh-bootloader/scripts/acceptance-tests IAAS
  ```

3. Add tests

Add unit tests for your feature.

## Terraform Template Changes

If you made any changes to the terraform templates, you need to run this script
to update the generated `templates.go` file.

```
~/bosh-bootloader/scripts/update_terraform_templates <IaaS>
```

Pass in the IaaS that needs updating as the argument (i.e. aws, gcp, azure)

## Vendor a dependency

We are currently using `dep` to vendor our dependencies - submodules used to be what we had in the past.
If you need to add a dependency to the vendor directory ie you imported some new library code just run:

  ```sh
    dep ensure github.com/some-user/your-repo
  ```

## Need help?

The Infrastructure team is available in the #bbl-users channel in [CF slack](https://cloudfoundry.slack.com).
