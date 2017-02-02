### Summary

This guide is a walkthrough for deploying a BOSH director with `bbl`
on AWS. Upon completion, you will have the following:

1. A BOSH director instance in the availability zone of your choice.
1. A set of randomly generated BOSH director credentials.
1. A generated keypair allowing you to SSH into the BOSH director and
any instances BOSH deploys.
1. A copy of the manifest the BOSH director was deployed with
1. A basic cloud config

### Preparing your environment

`bbl` requires `bosh-init` to deploy the BOSH director. Head over to bosh.io
for [installation instructions](http://bosh.io/docs/install-bosh-init.html).

To install `bbl` go to the
[releases page](https://github.com/cloudfoundry/bosh-bootloader/releases/latest)
and download the latest version for your platform.

To install move bbl into your PATH.

For Mac OS X/Linux machines you can do the following:

```
$ chmod +x ~/Downloads/bbl-*
$ sudo mv ~/Downloads/bbl-* /usr/local/bin/bbl
```

### Creating an IAM user

In order for `bbl` to interact with AWS, an `IAM` user must be created.
This user will be issuing API requests to create the infrastructure such
as EC2 instances, load balancers, subnets, etc.

The user must have the following `policy`:
```
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Effect": "Allow",
            "Action": [
                "ec2:*",
                "cloudformation:*",
                "elasticloadbalancing:*",
                "iam:*"
            ],
            "Resource": [
                "*"
            ]
        }
    ]
}
```

To create a user and associated policy with the AWS CLI run the 
following commands:
```
$ aws iam create-user --user-name "bbl-user"
$ aws iam put-user-policy --user-name "bbl-user" \
	--policy-name "bbl-policy" \
	--policy-document "$(pbpaste)"
$ aws iam create-access-key --user-name "bbl-user"
```

The `create-access-key` command will write an "access key id" and "secret 
access key" to the terminal. These values are important and should
be kept secret. In the next section `bbl` will use these commands to
create infrastructure on AWS.

### Creating infrastructure and BOSH director

`bbl` will create infrastructure and deploy a BOSH director with the
following command:

```
bbl up \
	--aws-access-key-id <INSERT ACCESS KEY ID> \
	--aws-secret-access-key <INSERT SECRET ACCESS KEY> \
	--aws-region us-west-1 \
	--iaas aws
```

The process takes around 5-8 minutes. When the process is finished
a file named `bbl-state.json` will be created in the current working
directory. This file is very important as it contains credentials
and other metadata related to your BOSH director and infrastructure.
It is highly recommended that you backup this file into version control
or another safe location. For more info about the `bbl-state.json` see
the "State management" section.

### State management

The `bbl-state.json` is an important file that contains confidential
information about your infrastructure. 

The state file allows you to easily upgrade to the newest BOSH director 
and stemcell versions when new versions of bbl are released. It will
allow you to issue a `bbl destroy` to destroy the many resources that 
`bbl` creates in AWS which is much easier and less error prone
than manually finding the resources in the AWS console or CLI. 

Backing up this file into a safe place is highly recommended. The file 
should never be modified by hand.

`bbl-state.json` contains the following:

- AWS access key ID, secret access key, region
- CloudFormation stack name
- Private key (for accessing EC2 instances BOSH deploys)
- Environment ID (unique ID for tag on all resources bbl deploys)
- BOSH director username and password
- BOSH director IP
- BOSH director SSL CA, certificate, private key
- bosh-init state and manifest (for the currently deployed director)

The best way to extract this info is by issuing commands like 
```
$ bbl director-username
some-username
$ bbl director-ca-cert
--- BEGIN CERTIFICATE ---
...
--- END CERTIFICATE ---

and so on...
```

In order to run these commands, the current directory must contain the
`bbl-state.json`.

### Connecting to the BOSH director

To setup your BOSH CLI with the new director you'll need the following
commands to get the credentials:

```
$ bbl director-address
https://23.248.87.5:25555
$ bbl director-username
user-d3783rk
$ bbl director-password
p-23dah71skl
$ bbl director-ca-cert
-----BEGIN CERTIFICATE-----
MIIDtzCCAp+gAwIBAgIJAIPgaUgWRCE8MA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
...
-----END CERTIFICATE-----
```

To login:

```
$ bosh target 23.248.87.55 <INSERT TARGET NAME>
Username: user-d3783rk
Password: p-23dah71sk1
```

Display cloud config:
```
$ bosh cloud-config
...
```

Now you're ready to deploy software with BOSH!
