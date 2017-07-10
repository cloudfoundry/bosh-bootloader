# Contributing to BOSH-Bootloader

We're glad you are here - if you are reading this you are one step closer to submitting a PR to bbl (bosh-bootloader).
There are just a few things to keep in mind before you send over a PR.

## Run low-level tests

```
  ~/bosh-bootloader/scripts/test
```

## Run acceptance tests

```
  ~/bosh-bootloader/scripts/acceptance-tests IAAS
```

## Add tests

Add unit tests (and acceptance test) to try out your feature.

## Vendor a dependency

We are currently using `dep` to vendor our dependencies - submodules used to be what we had in the past.
If you need to add a dependency to the vendor directory ie you imported some new library code just run:

```sh
  dep ensure github.com/some-user/your-repo
```

## Need help?

Your friendly everyday bbl'ers are available to help you on [slack](https://cloudfoundry.slack.com) in
the #bbl-users channel.
