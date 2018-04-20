## iso-segs-gcp

Creates a single routing isolation segment on GCP, including dedicated load balancers and firewall rules.

```
cp -r bosh-bootloader/plan-patches/iso-segs-gcp/. some-env/
bbl up
```

Disclaimer: this is a testing/development quality patch.  It has not been subject to a security review -- the firewall rules may not be fully locked down.
Please don't run it in production!
