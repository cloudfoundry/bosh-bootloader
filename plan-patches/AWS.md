## AWS Plan Patches

Plan patches can be used to customize the IAAS
environment and bosh director that is created by
`bbl up`.

A patch is a directory with a set of files
organized in the same hierarchy as the bbl-state dir.

* <a href='#iso-segs-aws'>Add Isolation Segments</a>
* <a href='#tf-backend-aws'>Use S3 Bucket for Terraform State</a>
* <a href='#iam-profile-aws'>Provide IAM Instance Profile for BOSH Director</a>
* <a href='#acm-aws'>Use ACM for Load Balancer Certs</a>

## <a name='iso-segs-aws'></a> iso-segs-aws

To create an isolation segment on aws, the files in this directory
should be copied to your bbl state directory.

The steps might look like such:

```
mkdir banana-env && cd banana-env

bbl plan --name banana-env --lb-type cf --lb-cert lb.crt --lb-key lb.key

cp -r bosh-bootloader/plan-patches/iso-segs-aws/. .

TF_VAR_isolation_segments="1" bbl up
```


## <a name='tf-backend-aws'></a> tf-backend-aws
Stores the terraform state in a given bucket on Amazon S3.

```
cp -r bosh-bootloader/plan-patches/tf-backend-aws/. .
```

Since the backend configuration is loaded by Terraform extremely early (before
the core of Teraform can be initialized), there can be no interplations in the backend
configuration template. Instead of providing vars for the bucket to an `s3_backend_override.tfvars`,
the values for the bucket name, region, and key for the state must be provided directly
in the backend configuration template.

Modify `terraform/s3_backend_override.tf` to provide the name and region of the bucket,
as well as the key to write the terraform state to.

Then you can bbl up.

```
bbl up
```


## <a name='iam-profile-aws'></a> iam-profile-aws

To use an existing iam instance profile on aws, the files in this directory
should be copied to your bbl state directory.

The steps might look like such:

```
mkdir banana-env && cd banana-env

bbl plan --name banana-env

cp -r bosh-bootloader/plan-patches/iam-profile-aws/. .
```

Write the name of the iam instance profile in `vars/iam.tfvars`.

```
bbl up
```

Providing the iam instance profile the bosh director means that the iam policy for
the user you give to bbl does not require `iam:*` permissions.


## <a name='acm-aws'></a> acm-aws

This is a patch for using AWS Certificate Manager to issue load balancer TLS certs.

First you're going to need a Route53 Zone for your system domain. ACM will verify that you own the domain
that it's going to produce certs for. This must be done prior to bbl'ing up.

1. Pick your system domain, then go to the AWS console and create a hosted zone to match.

1. Make sure your registrar or parent domain points to the name servers in the new hosted zone's default NS record.

1. Follow the normal steps for a plan patch: Invoke 
   ```
	 mkdir berry-env && cd berry-env
   bbl plan --lb-type cf  --lb-domain the.route53.zone.you.just.made.com \
            --lb-cert ../certs/fake.crt --lb-key ../certs/fake.key
   cp -r bosh-bootloader/plan-patches/acm-aws/. .
	 bbl up
   ```
   you'll need to provide certs to bbl, but they won't end up used, so you should be fine with empty or garbage files.	

1. Once you've bbl'd up, you should be able to deploy a cf and it'll have working, ACM managed certificates, provided that you
   supply cf-deployment with `-v system_domain=the.route53.zone.you.just.made.com`

1. Note, at the time of writing this plan-patch, there are issues in the terraform aws provider that prevent us from
   successfully bbl'ing down on the first try. If you see the issue described in https://github.com/terraform-providers/terraform-provider-aws/issues/3866 , just bbl down again.
