This directory contains a `Makefile.example` that you should use as the starting point for your own `Makefile`.

Assuming that you want to store all bbl artefacts in a directory named `iaas`:

1. Copy the `Makefile.example` from this directory to `iaas/Makefile`
1. Edit the `### REQUIRED VARS ###` section in `iaas/Makefile` to match your environment
1. Run `make`

To preview any make target (a.k.a. dry-run), run e.g. `make deploy_cf -n`

On OS X, it is recommended to use GNU Make >= v4, which you can install via `brew install make`.
