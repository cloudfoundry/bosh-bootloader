# BOSH Bootloader

Also known as `bbl` *(pronounced: "bubble")*, bosh-bootloader is a command line
utility for standing up a [CloudFoundry](https://cloudfoundry.org/) or [Concourse](https://concourse-ci.org) installation
on an IaaS.

`bbl` currently supports AWS, GCP, Microsoft Azure, Openstack and vSphere.

* [CI](https://infra.ci.cf-app.com/teams/main/pipelines/bosh-bootloader/)
* [Tracker](https://www.pivotaltracker.com/n/projects/1488988)

## What `bbl` does

![a list of steps that bbl executes, which are elaborated on below](theme/bbl-process.png)

### Generate terraform template
The first step that `bbl up` does is to generate a Terraform template based on your IAAS, IAAS region, and chosen load balancer type (or lack thereof).

The resulting Terraform template is emitted to the `terraform/bbl-template.tf` file within your state directory.

### Apply terraform template
After generating the Terraform template, `bbl up` will run Terraform to apply that template, using also a variables file located at
`vars/bbl.tfvars` within the state directory.

### Map terraform outputs to BOSH create-env vars
Having applied the Terraform template, we now have a number of Terraform outputs, such as subnet CIDRs, reserved IP addresses, and load balancer configuration.
`bbl` will transform those outputs into the inputs required by `jumpbox-deployment` and `bosh-deployment` and write them to the files `vars/jumpbox-vars-file.yml`
and `vars/director-vars-file.yml`.

### Execute BOSH create-env (jumpbox, director)
Next, `bbl` shells out to the BOSH CLI to run `bosh create-env` twice. The first time, `bbl` uses `jumpbox-deployment` and creates the jumpbox vm; the second time,
`bbl` uses `bosh-deployment` and creates the director VM. The exact commands that `bbl` will run are emitted to the `create-jumpbox.sh` and `create-director.sh`
files within the state directory.

### Generate cloud-config template
After the director VM comes up, `bbl` generates a base cloud-config, based on the IAAS, IAAS region, and chosen load balancer type.

### Map TErraform outputs to BOSH cloud-config vars
Having generated a base cloud-config template, `bbl` maps Terraform outputs to cloud-config variables. These variables include network and subnetwork names,
security groups or tags, and CIDR ranges, as well as load balancer target pool names.

### Update cloud-config (director)
Finally, `bbl` will update the director's cloud config, by shelling out to `bosh update-cloud-config`.
