# CloudFoundry installation
---

## Steps

### Starting point

This manual assumes that you have installed the [Dependencies](README.md#install-dependencies) mentioned in the main `README.md` on either your local machine or your `bastion server`/`jumpbox` in AWS EC2.

### Additional dependencies

- Make sure you did `gem install rake io-console bundler json`
- spiff ([installation instructions](https://github.com/cloudfoundry-incubator/spiff#installation))

### Actual CloudFoundry installation

1. Run `bbl unsupported-deploy-bosh-on-aws-for-concourse` which will give you a fully functional BOSH director and AWS infrastructure, created via an Amazon CloudFormation Stack.
2. Run `bbl director-address`, `bbl director-username` and `bbl director-password` to obtain the BOSH director address, username and password.
3. [Target the BOSH Director](https://docs.cloudfoundry.org/deploying/common/create_a_manifest.html#target) with your BOSH CLI.
4. Create an Amazon RDS PostgreSQL database in the `bbl` VPC and the `bbl-aws-*-InternalSecurityGroup-*` Security Group, make sure to create the `ccdb` and `uaadb` databases and [install the required PostgreSQL database extensions](https://docs.cloudfoundry.org/deploying/aws/cf-stub.html#editing) to be available for CloudFoundry: `uuid-ossp`, `pgcrypto`, and `citext`.
4. [Record the BOSH Director UUID](https://docs.cloudfoundry.org/deploying/common/create_a_manifest.html#uuid) and then [Create a Deployment Manifest Stub](https://docs.cloudfoundry.org/deploying/common/create_a_manifest.html#create-stub) via the [IaaS-specific instructions for AWS](https://docs.cloudfoundry.org/deploying/aws/cf-stub.html).
5. [Clone the cf-release GitHub Repository](https://docs.cloudfoundry.org/deploying/common/create_a_manifest.html#clone) and [Generate the (compiled) Manifest](https://docs.cloudfoundry.org/deploying/common/create_a_manifest.html#generate-manifest).
6. Ensure that the [recommended BOSH stemcell](http://bosh.io/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent) for your `cf-release` is uploaded to the director, as well as the [cf-release version](http://bosh.io/releases/github.com/cloudfoundry/cf-release?all=1) itself.
7. Target the correct deployment for BOSH with the name specified in your Deployment Manifest Stub: `bosh deployment cf-deployment`.
8. Run `bosh deploy` and wait patiently. Make sure your [EC2 Resource Limit](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-resource-limits.html) allows you to spin up the amount of EC2 instances you want here.
9. If all went well, you need to either acquire a valid SSL certificate, key, and optional chain or [generate self-signed ones (for testing purposes)](http://www.akadia.com/services/ssh_test_certificate.html).
10. Run `bbl unsupported-create-lbs --type cf` to create the ELBs needed by the gorouter and diego SSH proxies, which will update your Amazon CloudFormation Stack.
11. Run `bbl lbs` to get the Amazon ELB endpoint, then make sure to [configure the DNS for your domain(s)](https://docs.cloudfoundry.org/devguide/deploy-apps/routes-domains.html#domains-dns).
12. [Add the elb(s)](https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/42#issuecomment-230710972) to the Deployment Manifest Stub's `resource_pool` `cloud_properties` for your `routers`.
13. [Add an additional security group](https://github.com/pivotal-cf-experimental/bosh-bootloader/issues/42#issuecomment-229110469) to the (compiled) Manifests' `resource_pool` `cloud_properties` for your `routers`, this additional security group allows the ingress from our ELB that's required and is created by the previous `bbl` command.
14. [Verify the Deployment](https://docs.cloudfoundry.org/deploying/common/deploy.html#verify) and enjoy your fresh CloudFoundry installation!
