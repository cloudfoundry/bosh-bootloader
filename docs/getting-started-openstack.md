# Getting Started: OpenStack

This guide is a walkthrough for deploying a BOSH director with `bbl`
on OpenStack. Upon completion, you will have the following:

1. A BOSH director
1. A jumpbox
1. A set of randomly generated BOSH director credentials
1. A generated keypair allowing you to SSH into the BOSH director and
any instances BOSH deploys
1. A copy of the manifest the BOSH director was deployed with
1. A basic cloud config

`bbl` creates and maintains the lifecycle of the OpenStack infrastructure, jumpbox and BOSH director.  
It creates a network, a router, a floating IP, a key pair, security groups, and
security group rules on OpenStack.  
Prerequisite is that an OpenStack project, external network, and user already
exist.

## Create a Jumpbox and a BOSH Director

1. Export environment variables.
    ```
    export BBL_IAAS=openstack
    export BBL_OPENSTACK_AUTH_URL=
    export BBL_OPENSTACK_AZ=
    export BBL_OPENSTACK_NETWORK_ID=  # external network ID
    export BBL_OPENSTACK_NETWORK_NAME=  # external network name
    export BBL_OPENSTACK_PASSWORD=
    export BBL_OPENSTACK_USERNAME=
    export BBL_OPENSTACK_PROJECT= # same as tenant
    export BBL_OPENSTACK_DOMAIN=
    export BBL_OPENSTACK_REGION=

    # optionally
    #export BBL_OPENSTACK_CACERT_FILE= # custom CA certificate when communicating over SSL; either path to file or contents of certificate
    #export BBL_OPENSTACK_INSECURE=  # e.g. "true", default: "false"
    #export BBL_OPENSTACK_DNS_NAME_SERVERS=  # e.g. "8.8.8.8,9.9.9.9", default: "8.8.8.8"
    ```

1. Create OpenStack resources, jumpbox and bosh director.
    ```
    bbl up
    ```

## Next Steps

* [Target the BOSH Director](howto-target-bosh-director.md)
