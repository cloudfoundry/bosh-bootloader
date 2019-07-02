## n-cf-gcp

By default, bbl sets up environment for one single CF deployment.
This plan patch adds support for multiple CF environments.

Details:
It creates separate load balancers and DNS records for each of the environments.
The domain names for each environments are derived from your `lb-domain`
flag (e.g. cf0.banana-env.com, cf1.banana-env.com etc).

Important: To deploy cf you will need to use the `cf-lb-ops.yml` in order to
properly setup the load balancers.

Example
```
bosh -d cf0 deploy $CF_D/cf-deployment.yml \
  -o $CF_D/operations/rename-network-and-deployment.yml \
  -v deployment_name=cf0 \
  -v network_name=default \
  -o cf-lb-ops.yml \
  -v env=0 \
  -v system_domain=cf0.banana-env.com
```

The steps might look like such:

```
export TF_VAR_cf_env_count=<number-of-cf-deployments>
mkdir banana-env && cd banana-env

bbl plan --name banana-env --lb-type=cf --lb-cert=./certs/fake.crt --lb-key=certs/fake.key --lb-domain=banana-env.com

cp -r bosh-bootloader/plan-patches/n-cf-gcp/* .

bbl up
./update-configs.sh
```
