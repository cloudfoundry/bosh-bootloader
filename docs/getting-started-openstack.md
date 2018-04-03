# Getting Started: OpenStack

## Jumpbox, Director

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
1. Create jumpbox and bosh director.
    ```
    bbl up
    ```
