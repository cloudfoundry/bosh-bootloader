# Migration Guide

## bbl v5 -> bbl v6

Known issue: trying to destroy a `bbl5` environment with `bbl6` requires you to
`bbl6 plan && bbl6 up` against that state directory.

This is due to a PR released in `bbl6` that forwarded terraform outputs directly
through to the bosh deployment vars files for `create-env` and `delete-env`.

By running `bbl6 up`, the terraform outputs are updated and so too are the
bosh deployment vars files.

```
bbl5 up

# some time passes

bbl6 plan
bbl6 up

bbl6 destroy
```
