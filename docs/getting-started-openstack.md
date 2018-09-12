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

`bbl` creates and maintains the lifecycle of the jumpbox and BOSH director.

It does not create any networks, security groups, or load balancers on OpenStack.

## Create a Jumpbox and a BOSH Director

1. Export environment variables.
    ```
    export BBL_IAAS=openstack
    export BBL_OPENSTACK_INTERNAL_CIDR=
    export BBL_OPENSTACK_EXTERNAL_IP=
    export BBL_OPENSTACK_AUTH_URL=
    export BBL_OPENSTACK_AZ=
    export BBL_OPENSTACK_DEFAULT_KEY_NAME=
    export BBL_OPENSTACK_DEFAULT_SECURITY_GROUP=
    export BBL_OPENSTACK_NETWORK_ID=
    export BBL_OPENSTACK_PASSWORD=
    export BBL_OPENSTACK_USERNAME=
    export BBL_OPENSTACK_PROJECT=
    export BBL_OPENSTACK_DOMAIN=
    export BBL_OPENSTACK_REGION=
    export BBL_OPENSTACK_PRIVATE_KEY=
    ```

    ```powershell
    $env:BBL_IAAS="openstack"
    $env:BBL_OPENSTACK_INTERNAL_CIDR=
    $env:BBL_OPENSTACK_EXTERNAL_IP=
    $env:BBL_OPENSTACK_AUTH_URL=
    $env:BBL_OPENSTACK_AZ=
    $env:BBL_OPENSTACK_DEFAULT_KEY_NAME=
    $env:BBL_OPENSTACK_DEFAULT_SECURITY_GROUP=
    $env:BBL_OPENSTACK_NETWORK_ID=
    $env:BBL_OPENSTACK_PASSWORD=
    $env:BBL_OPENSTACK_USERNAME=
    $env:BBL_OPENSTACK_PROJECT=
    $env:BBL_OPENSTACK_DOMAIN=
    $env:BBL_OPENSTACK_REGION=
    $env:BBL_OPENSTACK_PRIVATE_KEY=
    ```
1. Create jumpbox and bosh director.
    ```
    bbl up
    ```

## Next Steps

* [Target the BOSH Director](howto-target-bosh-director.md)