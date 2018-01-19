## Code walkthrough

- look at bbl/main.go - where do the user inputs go?
- look at the migrator - what are we migrating?

## WAT?!

- local variable called `globals` in ParseArgs
- Why does http still need `--skip-ssl-validation`? Why were we able to connect with http + `--skip-ssl-validation` but not https + `--skip-ssl-validation`

## Cheat sheet

- GIVEN: you want to use bbl for particular IAAS with sensible defaults and our team's credentials  
  THEN: run `eval_bbl_gcp_creds` or `eval_bbl_aws_creds` or `eval_bbl_azure_creds` (requires lpass login)  
- GIVEN: you made changes to the terraform templates inside bbl and you want to see them in action with `bbl`  
  THEN run `./scripts/update_terraform_templates CURRENT_IAAS` from the bosh-bootloader directory  
- GIVEN: you made changes to the bbl code and you want to see them in action with `bbl`  
  THEN: run `./scripts/bbl` from the bosh-bootloader directory  
- GIVEN: you bbled up an lb environment you would like to be able to reach with DNS  
  THEN: create a new NS record in the 'infrastructure' zone in Cloud DNS. Give it the list of name servers found in the NS record in the zone for your environment.  

## Iaas projects
- aws  
  => CF Infrastructure: for development and test  
  => CF Infrastructure Acceptance: for PM story acceptance  

- gcp  
  => CF-Infrastructure: development playground  
  => CF-Infrastructure-BBL-Test: for jobs in the bosh-bootloader pipeline that bbl up and bosh deploy  
  => CF-Infrastructure-CI: for jobs in other pipelines besides bosh-bootloader like consul  

- azure  
  => CF-Infrastructure: for everything

- vsphere
  => Pizza boxes/khaleesi: for everything

## Connor

- What kind of context helps you the most? i.e. Do you approach domains from the user's perspective or more technical bottom up?

Vocabulary is a big deal for me, followed by boxes-and-lines architecture. Definitely technical, but top-down.

- Describe your favorite project youve worked on

Distributed camera project
Big challenging concurrency problems.
Performance problems.
Greenfielding and trail-blazing over cleaning up a mess
Refactoring when there is a good test suite

- Describe your least favorite project

Content identification system:
Hand-rolled infrastructure shenanigans
Testing in prod
Stressful

- What aspects of cf infrastructure are you most familiar with? bosh? gcp? vsphere? consensus algorithms?

Diego circa 2016
Bosh pre-links pre-credhub
First release of bbl
AWS
Concourse

- What aspects are you unfamiliar with?

Never used gcp, vsphere, azure
Never used terraform


## Day 1 Exercise: Deploy cf using bbl

If you come across any terminology you aren't familiar with then add it to the Vocabulary section below.

1. from the bosh-bootloader project directory

1. cd ignored

1. mkdir day-one

1. cd day-one

1. run `eval_bbl_gcp_creds`

1. Run `bbl plan --name day-one --iaas gcp --lb-type cf --lb-cert ../fake_cert_stuff/fake.crt --lb-key ../fake_cert_stuff/fake.key --lb-domain day-one.infrastructure.cf-app.com`
  - Run `tree .` and look around at the files that were created. How are they organized?
  - What files look like they are related to terraform? Which ones look like they are for bosh?
  - Look around in the [cf-infrastructure project on gcp](https://console.cloud.google.com/home/dashboard?project=cf-infra). There should be nothing with day-one in the name.

1. Run `bbl up`
  - What changed in the file system?
  - Look around in the vars directory. Where are all the credentials?
  - From the /vars directory run `terraform output`
  - Look around in the [cf-infrastructure project on gcp](https://console.cloud.google.com/home/dashboard?project=cf-infra). Which resources have day-one in the name? How are they related to each other?

1. Create a new set of NS records in the infrastructure zone with the name servers from the day-one zone.

1. Run `bbl print-env`. What does it do?

1. Run `eval "$(bbl print-env)"

1. Run `bosh upload-stemcell https://bosh.io/d/stemcells/bosh-google-kvm-ubuntu-trusty-go_agent\?v\=3468.15`

1. Change to the `cf-deployment` directory.

1. Run `bosh -d cf deploy cf-deployment.yml --vars-store ~/go/src/github.com/cloudfoundry/bosh-bootloader/ignored/day-one/vars/cf-deployment-vars.yml -v system_domain="day-one.infrastructure.cf-app.com" -o operations/use-compiled-releases.yml -n`

1. Look around in the [cf-infrastructure project on gcp](https://console.cloud.google.com/home/dashboard?project=cf-infra). Which resources did bosh create? How do they relate to the resources that `bbl up` created?

1. Clean up all the resources you delployed. Try the gcp console and [leftovers](git@github.com:genevievelesperance/leftovers). Delete resources in approximately this order:
  - instances
  - firewall rules
  - health checks
  - load balancers
  - instance groups
  - external ips
  - dns records in your zone
  - your zone
  - your vpc network
  - disks
  - images

## Pipelines

1. bosh-bootloader:

1. terraforming:

1. consul:

1. bosh-deployment:

1. socks5-proxy:

1. bbl-latest:

1. infrastructure-ci:

1. check-a-record:

1. dockerfiles:

1. destiny:

1. bosh-test:

1. gomegamatchers: gomegamatchers is the infrastructure team's extension of gomega for ginko.

## Mission of the BBL team

1-click install?
Solving the bootstrapping problem once and for all...?
Reduce toil for other CF teams

## Users of BBL

100% CF teams use it
Open-source tire-kicker
Power users (mostly other CF teams)
Operators in PCFS (coming)

## Vocabulary

- environment: all the iaas usually referring to the vpc. the environment name is either user-provided or generated on bbling up and is used to nam

- ops-file: bosh's yaml munging concept. bosh with interpolate a yml file by applying ops-files to it.

- HCL: Hashicorp Configuration Language. We use it for terraform.

- BBL State Dir: the directory where you ran `bbl plan` and `bbl up`. It holds all the variables and credentials for the environment.

- BOSH vars-file: inputs to the bosh interpolation. These are created from the terraform outputs, in most cases passed through directly.

- BOSH vars-store: outputs from running bosh create-env. mostly/all credentials.

- socks5: a traditional, secure, socket proxy. We set up a socks5 proxy when you run `gobosh_target` aka `eval "$(bbl print-env)`

## Unanswered questions

- Why can't bbl add the DNS records for infrastructure.cf-app.com

## Day One Survey - Connor

Welcome, new team member! Thanks for taking time to fill out our ramp-up survey. Bringing new people onto the project is an important opportunity, because unlike your teammates, you haven’t acclimated to the weird parts of the project and codebase. This means that we have a brief window before you acclimate to shed light on the parts of the project that need cleaning up.
To help us with this process, any time you find something weird, or difficult to understand, please say something! That way we can improve things for the next person to join.
We’re also experimenting with collecting data from projects around the office, to identify ways that teams can improve ramp up in the office as a whole. You can share as many or as few of your answers as you’d like for that purpose; talk to your anchor for more.

- What did you work on today?

re-bbl-upping a patch allowing bbl to deploy a smaller number of instance groups, generic onboarding tasks

- What about today was the biggest help to your contributing to the project?

learning, getting access to resources

- What about today was the biggest obstacle to your contributing to the project? 

needing to send a ton of ask tickets for accounts and so on

- Of the parts of your project you’ve seen so far, what is the most confusing?

haven't seen enough to feel real confused yet... the stuff we looked at around the routing layer had me confused

- After your experience today, what is a question you wish you knew the answer to?

I feel like I could use a referesher on how we configure our lbs+routers+tls termination
