## acm-aws

This plan patch is no longer supported, bbl now uses Network Load Balancers on AWS.
They don't support ACM to issue load balancers. For this to work you'd need to modify
this plan patch to override the Load Balancers to use Application Load Balancers.

This is a patch for using AWS Certificate Manager to issue load balancer TLS certs.

First you're going to need a Route53 Zone for your system domain. ACM will verify that you own the domain
that it's going to produce certs for. This must be done prior to bbl'ing up.

1. Pick your system domain, then go to the AWS console and create a hosted zone to match.

1. Make sure your registrar or parent domain points to the name servers in the new hosted zone's default NS record.

1. Follow the normal steps for a plan patch: Invoke 
   ```
   export BBL_SOURCE=${GOPATH}/src/github.com/cloudfoundry/bosh-bootloader/
	 mkdir berry-env && cd berry-env
   bbl plan --lb-type cf  --lb-domain the.route53.zone.you.just.made.com \
            --lb-cert ../certs/fake.crt --lb-key ../certs/fake.key
   cp -r ${BBL_SOURCE}/plan-patches/acm-aws/. .
	 bbl up
   ```
   you'll need to provide certs to bbl, but they won't end up used, so you should be fine with empty or garbage files.	

1. Once you've bbl'd up, you should be able to deploy a cf and it'll have working, ACM managed certificates, provided that you
   supply cf-deployment with `-v system_domain=the.route53.zone.you.just.made.com`

1. Note, at the time of writing this plan-patch, there are issues in the terraform aws provider that prevent us from
   successfully bbl'ing down on the first try. If you see the issue described in https://github.com/terraform-providers/terraform-provider-aws/issues/3866 , just bbl down again.

