## v7.0.0 (Unreleased)

**BACKWARD INCOMPATIBILITIES / NOTES:**

**FEATURES / IMPROVEMENTS:**

**BUG FIXES:**

## v6.0.0 (January 26, 2018)

**BACKWARD INCOMPATIBILITIES / NOTES:**
* Deprecated commands were removed: `create-lbs`, `update-lbs`, `delete-lbs`, `cloud-config`, `bosh-deployment-vars`, `jumpbox-deployment-vars`
* Deprecated flags were removed: `--no-director`, `--ops-file`, `--gcp-project-id`, `--gcp-zone`
* Requies `bosh-cli v2.0.48`

**FEATURES / IMPROVEMENTS:**
* Support to read cloud-config ops-files alphabetically when those files are included in the cloud-config directory of your bbl state directory.
* Support to apply overrides and additional terraform templates when those files are included in the terraform directory of your bbl state directory.
* Support for Azure concourse load balancers.
* IAAS credentials no longer written to bbl state directory.
* Variables written to `vars/terraform.tfvars` with a `jumpbox__` prefix will
be passed through to the deployment vars file for the jumpbox.
* Variables written to `vars/terraform.tfvars` with a `director__` prefix will
be passed through to the deployment vars file for the director.
* Provide your own `.tfvars` file in `vars/`
* In order to avoid having your changes overwritten during upgrades, bbl will
now use `create-director-override.sh` and `create-jumpbox-override.sh` if provided instead of the defaults.
* Bring your own VPC
* Added `BBL_STATE_DIRECTORY` environment variable.

**BUG FIXES:**
* vSphere cloud-config reserves the IPs for the jumpbox and director.


## v5.1.0 (October 18, 2017)
v5.0.0 was a pre-release for the major breaking changes.

**BACKWARD INCOMPATIBILITIES / NOTES:**
* bbl no longer remembers your IaaS keys for you. This is another **breaking**
change that affects all IaaSes. Please set IaaS credentials for any command
that will modify Infrastructure (up, down, create-lbs, etc.)
* If you want to **migrate from bbl 3.0 to 5.0 on AWS** environments, you must hit a
`4.0.0` release first. `4.0.0` migrates you from a CloudFormation backend, which is the
IaaS specific past, to a Terraform backend, which is the nearly-IaaS-agnostic future.
* If you **ssh to your job vms** in your pipeline using bbl director-address, please
adjust to use bbl jumpbox-address. This is available in version `4.10.0` and `4.10.1`
in a forward compatible manner: it always returns the `BOSH_GW_HOST`, whether the
director is the gateway or the jumpbox is the gateway.
* If you **ssh to your BOSH director** in your pipeline, well, you can't very easily
now. With SOCKS5 proxies we set up the gateway for you, and all BOSH commands
respect `BOSH_ALL_PROXY` (like bosh ssh).
We've [documented](https://github.com/cloudfoundry/bosh-bootloader/blob/master/docs/howto-ssh.md)
how you can SSH to the director if you really need to.

**FEATURES / IMPROVEMENTS:**
* BOSH gets deployed with credhub now. This should be a nonbreaking change,
as long as you specify a var store when you deploy it will continue to work as you expect.
* The BOSH director API is now only accessible via a jumpbox. This is the most
painful and most **breaking** part of this release. bbl will set up a SOCKS5 proxy
for you if you run `eval "$(bbl print-env)"`. This comes for free if you use
cf-deployment-concourse-tasks or bosh-deployment-resource.
* Adds a credhub Loadbalancer and DNS for the Secure Service Instance Credential effort.
* Adds a checksum to the release notes
* Opens the BOSH credhub to traffic from the jumpbox

**BUG FIXES:**
* Fixes jumpbox related commands on Azure that were broken in `5.0.0`.


## v4.0.0 (July 28, 2017)

**BACKWARD INCOMPATIBILITIES / NOTES:**
* BBL 4.0 is a breaking change, but only for AWS environments. While I emphasize
the POTENTIALLY BREAKING part we have gone to great lengths to automatically
migrate users and test this migration progress to make it as smooth as possible
while we migrate you from cloudformation to Terraform. This migration will happen
the next time you attempt to modify your infrastructure using 4.0 or above.
* AWS users will however experience downtime on load-balanced services (gorouter
or atc). If you have a non-ephemeral environment with uptime concerns, open an
issue and we can help suggest a zero downtime migration.

To migrate and restore load balancing teams must:
1. Install terraform if you haven't already
1. run bbl create-lbs again (or just bbl up if you don't use load balancers).
1. `bosh recreate router`, `atc`, and other vms that use the load balancer VM extensions (or ignore this step if you don't use load balancers)

**FEATURES / IMPROVEMENTS:**
* IAM instance profiles. Fewer IaaS credentials on fewer disks.
* The full suite of loadbalancers that GCP already supports
* Error messages without opening up AWS console to look at CloudFormation
* A DNS zone with the names of these loadbalancers if they use bbl create-lbs --domain.
* A lot of stealth-mode changes to the --jumpbox feature flag

**BUG FIXES:**


## v3.0.0 (March 15, 2017)

**BACKWARD INCOMPATIBILITIES / NOTES:**
* This is a breaking change which was necessary to roll out new directors built
using bosh-deployment. If you attempt to use an existing state directory created
with bbl v2 then bbl will refuse to proceed. For ephemeral environments this
shouldn’t pose a problem, just make sure you have finished teardown before upgrading.
For long-lived environments you may want to stand up a new environment before
tearing down the old with bbl v2.4.1.
* We’ve added a requirement that the bosh 2.0+ CLI be installed on your path and
be named “bosh”. bosh-init is no longer required.

**FEATURES / IMPROVEMENTS:**
* For trivial preference changes, such as the number of workers or ARP cache
flush, you can supply an ops file with bbl up -o ops-file.yml.
* For more intense changes such as boshlite-on-gcp or credhub you may find it
easier to run bbl up --no-director and then deploy your director with
bosh-deployment. You can read more about this workflow in our docs.

**BUG FIXES:**

