# 2. Replace go-bindata with packr2

Date: 2019-02-22

## Status

Accepted

## Context

The issue motivating this decision, and any context that influences or constrains the decision.

The original author of go-bindata delete their account and then the repo was
recreated under a different owner. The dependency has shifted around enough that
we have lost faith in the intention of the maintainers

[more details here](https://twitter.com/francesc/status/961249107020001280?lang=en)

Also, some of the development use cases around go-bindata (like what is bundled into the code
during a test run or final build) made it hard to reason about.

## Decision

Use [Packr2](https://github.com/gobuffalo/packr/tree/master/v2) instead.

## Consequences

- Improved security, reduced risk of a potential bad actor owning go-bindata

- It improves developer experience

  If you use `go {build,install}`, it will load the files off disk

-	You can no longer use `go get ./...` to install bbl. If you install
  bosh-bootloader, via `go get ./...`, the build will fail since
  the dependencies (the terraform binary in terraform/binary\_dist/ and
  jumpbox-deployment and bosh-deployment in bosh/deployments/) do not exist yet.

  You need to manually add these.

- There is no longer a dev-shim and testing now can fake out the content that
  was previously provided by said shim

- Builds that want to actually bundle the source into bbl must use the packr2
  build tool
