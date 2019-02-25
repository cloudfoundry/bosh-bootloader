## Why?

This directory is now here to help with the usage of [packr2](https://godoc.org/github.com/gobuffalo/packr/v2)

When building the actual bbl binary, this is where your jumpbox / bosh
deployment repos should end up so they can be included in the final build.

## What happens if this directory isn't empty

Whatever jumpbox / bosh deployment repos you have residing in this directory
will be built into the bbl. Use caution.

## Submodules?

We have submoduled in both deployments as they are pretty tightly coupled to a
particular version of bbl. This should help when building a one-off (you can
still checkout the directories to whatever sha you want)
