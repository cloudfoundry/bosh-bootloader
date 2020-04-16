## Network Load Balancers on GCP (AKA Regional Load Balancers)

Use this plan patch to replace the Global HTTP Load Balancer for CF HTTP(S)
traffic with a Network Load Balancer (which is a regional TCP Load Balancer).

Run `bbl plan` with `--lb-type cf` and `--lb-domain <your-domain>`. Then copy the plan patch into your bbl state directory before running `bbl up`:

```
cd your-bbl-state-dir
bbl plan --lb-type cf --lb-domain <some-domain> --lb-cert <some-cert> --lb-key <some-key>
cp -R bosh-bootloader/plan-patches/network-lb-gcp/* .
bbl up
```

Note: the cert and key used for `bbl plan` will not actually be used as the
TCP Load Balancer does not terminate TLS, but they still
need to be valid to make the initial `bbl plan` command succeed.
