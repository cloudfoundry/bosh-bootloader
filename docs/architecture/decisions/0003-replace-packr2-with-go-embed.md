# 2. Replace packr2 wiht go embed

Date: 21-02-2023

## Status

Inprogress

## Context

The issue motivating this decision, and any context that influences or constrains the decision.

The original author packr has discontinued the work on packr
and archived the repo https://github.com/gobuffalo/packr/ since Jun 25 2022
and as of go 1.16 `go:embed` was introduced see https://pkg.go.dev/embed

which is native file embedding feature of Go, or github.com/markbates/pkger.
It has an idiomatic API, minimal dependencies,
a stronger test suite (tested directly against the std lib counterparts),
transparent tooling, and more.


## Decision

Use [embed](https://pkg.go.dev/embed) instead.

## Consequences

- getting rid of an external tool packr2
