## Network Load Balancers on GCP (AKA Regional Load Balancers)

To replace the Global HTTP Load Balancer for CF HTTP(S) traffic with a Network Load Balancer (which is a regional TCP Load Balance), copy the plan patch into your bbl state directory before running `bbl up`:
```
cp -R bosh-bootloader/plan-patches/network-lb-gcp/* your-bbl-state-dir/
```
