## credhub-lb-aws

This patch will deploy an AWS network load balancer that forwards traffic to the Cloud Foundry runtime credhub that is used for [secure service credentials](https://github.com/pivotal-cf/credhub-release/blob/master/docs/secure-service-credentials.md)

You'll be able to reach credhub on the load balancer port 8844

1. Do the plan-patch dance:
   ```bash
   export BBL_SOURCE=${GOPATH}/src/github.com/cloudfoundry/bosh-bootloader/
   export DOMAIN=example.domain.com
	 mkdir cf-env && cd cf-env/
   bbl plan --name cf-env --lb-type cf --lb-cert ${CERT_PATH} --lb-key ${KEY_PATH} --domain ${DOMAIN}
   cp -r ${BBL_SOURCE}/plan-patches/credhub-lb-aws/. .
   bbl up
   ```
1. Once you've bbl'd up, deploy your cloudfoundry:
   ```bash
   git clone https://github.com/cf-deployment.git
   eval "$(bbl print-env)"
   bosh deploy -d cf cf-deployment/cf-deployment.yml \
     -o cf-deployment/operations/experimental/secure-service-credentials.yml \
     -o cf-deployment/operations/experimental/add-credhub-lb.yml \
     -o cf-deployment/operations/aws.yml \
     -v system_domain=${DOMAIN}
   ```

1. Wait a hot minute for your load balancer to find their targets, then open a new terminal to log in into credhub using credhub cli.
   If you use the same terminal you used to deploy CF, you'll be talking to the director credhub instead of the CF one.
   ```bash
   credhub login "http://$(bbl outputs | bosh int --path=/credhub_lb_url -):8844"
   ```
