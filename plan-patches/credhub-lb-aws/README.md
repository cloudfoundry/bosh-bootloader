## credhub-lb-aws

This patch will deploy an aws network load balancer that forwards traffic to a runtime credhub that can be used for securing service credentials https://github.com/pivotal-cf/credhub-release/blob/master/docs/secure-service-credentials.md

You'll be able to reach credhub on the load balancer port 8844

1. Do the plan-patch dance:
   ```bash
   export BBL_SOURCE=${GOPATH}/src/github.com/cloudfoundry/bosh-bootloader/
	 mkdir cf-env && cd cf-env/
   bbl plan --name cf-env
   cp -r ${BBL_SOURCE}/plan-patches/credhub-lb-aws/. .
   bbl up
   ```
1. Once you've bbl'd up, deploy your cloudfoundry:
   ```bash
   git clone https://github.com/cf-deployment.git
   bosh deploy -e cf-env -d cf cf-deployment/cf-deployment.yml \
     -o cf-deployment/operations/experimental/enable-instance-identity-credentials.yml \
     -o cf-deployment/operations/experimental/secure-service-credentials.yml \
     -o cf-deployment/operations/experimental/add-credhub-lb.yml
   ```

1. Wait a hot minute for your load balancer to find their targets, then login into credhub using credhub cli
   ```bash
   credhub login "http://$(bbl outputs | bosh int --path=/credhub_lb_url -):8844"
   ```