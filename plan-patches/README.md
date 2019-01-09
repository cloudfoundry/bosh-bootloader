## Plan Patches

Plan patches can be used to customize the IAAS
environment and bosh director that is created by
`bbl up`.

A patch is a directory with a set of files organized in the same hierarchy as the bbl-state dir.
`bbl down` won't delete files that bbl didn't create, and `bbl up` will make an effort to merge your terraform overrides and apply your jumpbox, director, and cloud-config ops files provided that they're in the right directories.

Many of these have additional prep steps or specific downstream bosh deployments in mind, so be sure to read the `README.md` of the patch you're trying to apply.

| Name | Purpose |
|:---  |:---     |
| **AWS** |     |
| [iam-profile-aws](iam-profile-aws/) | Provide IAM Instance Profile for BOSH Director |
| [acm-aws](acm-aws/) | Use Amazon Certificate Manager to issue load balancer certificates |
| [alb-aws](alb-aws/) | Use an Application Load Balancer instead of classic ELBs |
| [cfcr-aws](cfcr-aws/) | Deploy a CFCR with a kubeapi load balancer and aws cloud-provider |
| [iso-segs-aws](iso-segs-aws/) | Add Isolation Segments |
| [1-az-aws](1-az-aws/) | Only create resources in a single availability zone |
| [tf-backend-aws](tf-backend-aws/) | Store your terraform state in S3 |
| [prometheus-lb-aws](prometheus-lb-aws/) | Deploy a dedicated AWS network load balancer for your prometheus cluster |
| [s3-blobstore-aws](s3-blobstore-aws/) | Create S3 and IAM resources for an external blobstore |
| **GCP** |     |
| [bosh-lite-gcp](bosh-lite-gcp/) | For bosh-lites hosted on gcp |
| [cfcr-gcp](cfcr-gcp/) | Deploy a CFCR with a kubeapi load balancer and aws cloud-provider |
| [iso-segs-gcp](iso-segs-gcp/) | Add Isolation Segments |
| [byobastion-gcp](byobastion-gcp/) | From within a VPC, deploy a bosh director without a jumpbox |
| [tf-backend-gcp](tf-backend-gcp/) | Store your terraform state in GCS |
| [restricted-instance-groups-gcp](restricted-instance-groups-gcp/) | Create two seperate instance groups |
| [colocate-gorouter-ssh-proxy-gcp](colocate-gorouter-ssh-proxy-gcp/) | Helpful if you're trying to colocate everything |
| [prometheus-lb-gcp](prometheus-lb-gcp/) | Deploy a dedicated GCP load balancer for your prometheus cluster |
| **VSPHERE** |     |
| [cfcr-vsphere](cfcr-vsphere/) | Deploy a CFCR with a single master static IP and the vsphere cloud-provider |
| **OPENSTACK** |     |
| [cfcr-openstack](cfcr-openstack/) | Deploy a CFCR with a single master floating static IP |
| **Azure** |     |
| [cf-lite-azure](cf-lite-azure/) | Deploy a cf-lite on azure, one-box dev environment for CF |
| [cf-azure](cf-azure/) | Deploy a cf on azure |
| [cfcr-azure](cfcr-azure/) | Deploy a cfcr on azure |
