## HTTP2 load balancer for Azure

This is a patch that will enable http2 protocol in Azure's load balancer(Application Gateway) setup.

**Be aware** that requests to this load balancer might be incompatible with
older versions of gorouter. We recommend using this plan-patch only if you are
on a routing-release version with http2 enabled.

Compatible with routing-release version: x.x+ (will fill this in later) and
cf-deployment version: x.x+

To use, please follow the standard plan-patch steps.

```
mkdir your-env && cd your-env

bbl plan --name your-env

cp -r bosh-bootloader/plan-patches/http2-lb-azure/. .

bbl up
```
