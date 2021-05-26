## http2 load balancer for GCP

This is a patch that will enable http2 protocol for the backend services used in the GCP load
balancer.

This will override the existing HTTPS protocol that is configured on the load
balancer. **Be aware** that requests to the loadbalancer will upgrade its connection to http2. If
your gorouter is not up-to-date or does not allow http2 traffic, then the
requests to the gorouter will fail.

To use, please follow the standard plan-patch steps.

```
mkdir your-env && cd your-env

bbl plan --name your-env

cp -r bosh-bootloader/plan-patches/http2-lb-gcp/. .

bbl up
```
