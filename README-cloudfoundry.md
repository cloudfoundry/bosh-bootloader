# CloudFoundry installation
---

## Steps

### Starting point

This manual assumes that you have installed the Dependencies mentioned in `README.md` on either your local machine or your `bastion` server in AWS EC2.

### Actual CloudFoundry installation

1. Run `bbl unsupported-deploy-bosh-on-aws-for-concourse` which will give you a fully functional BOSH director and AWS infrastructure.
2. Run `bbl director-address`, `bbl director-username` and `bbl director-password` to obtain the BOSH director address, username and password.
3. [Target the BOSH Director](https://docs.cloudfoundry.org/deploying/common/create_a_manifest.html#target) with your BOSH CLI.
4. (optionally) Create an Amazon RDS PostgreSQL database in the `bbl` VPC and the `bbl-aws-*-InternalSecurityGroup-*` Security Group, make sure to create the `ccdb` and `uaadb` databases and install the required extensions to be available for CloudFoundry: `uuid-ossp`, `pgcrypto`, and `citext`.
4. [Record the BOSH Director UUID](https://docs.cloudfoundry.org/deploying/common/create_a_manifest.html#uuid) and then [Create a Deployment Manifest Stub](https://docs.cloudfoundry.org/deploying/common/create_a_manifest.html#create-stub) via the [IaaS-specific instructions for AWS](https://docs.cloudfoundry.org/deploying/aws/cf-stub.html).
5. [Clone the cf-release GitHub Repository](https://docs.cloudfoundry.org/deploying/common/create_a_manifest.html#clone) and [Generate the Manifest](https://docs.cloudfoundry.org/deploying/common/create_a_manifest.html#generate-manifest).
6. Ensure that the [recommended BOSH stemcell](http://bosh.io/stemcells/bosh-aws-xen-hvm-ubuntu-trusty-go_agent) for your `cf-release` is uploaded to the director, as well as the [cf-release version](http://bosh.io/releases/github.com/cloudfoundry/cf-release?all=1) itself.
7. Target the correct deployment for BOSH with the name specified in your `cf-stub.yml`: `bosh deployment cf-deployment`.
8. Run `bosh deploy` and wait patiently. Make sure your [EC2 Resource Limit](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/ec2-resource-limits.html) allows you to spin up the amount of EC2 instances you want.
