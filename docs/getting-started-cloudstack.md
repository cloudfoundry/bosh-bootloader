# Getting Started: CloudStack

This guide is a walkthrough for deploying a BOSH director with `bbl`
on Cloudstack. Upon completion, you will have the following:

1. A BOSH director
1. A jumpbox
1. A set of randomly generated BOSH director credentials
1. A generated keypair allowing you to SSH into the BOSH director and
any instances BOSH deploys
1. A copy of the manifest the BOSH director was deployed with
1. A basic cloud config

`bbl` creates and maintains the lifecycle of the CloudStack infrastructure, jumpbox and BOSH director.
It creates a network, a router, a floating IP, a key pair, security groups, and
security group rules on CloudStack.
Prerequisite is that an CloudStack project, external network, and user already
exist.

## Create a Jumpbox and a BOSH Director



1. Export environment variables.
    ```
    export BBL_IAAS=cloudstack
    export BBL_CLOUDSTACK_ENDPOINT=https://cloudstack.example.org/client/api
    export BBL_CLOUDSTACK_ZONE=MY-ZONE
    export BBL_CLOUDSTACK_SECRET_ACCESS_KEY=
    export BBL_CLOUDSTACK_API_KEY=
    ```

1. Create CloudStack resources, jumpbox and bosh director.
    ```
    bbl up
    ```

## Next Steps

* [Target the BOSH Director](howto-target-bosh-director.md)
