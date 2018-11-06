# Getting Started: vSphere

This guide is a walkthrough for deploying a BOSH director with `bbl`
on vSphere. Upon completion, you will have the following:

1. A BOSH director
1. A jumpbox
1. A set of randomly generated BOSH director credentials
1. A generated keypair allowing you to SSH into the BOSH director and
any instances BOSH deploys
1. A copy of the manifest the BOSH director was deployed with
1. A basic cloud config

`bbl` creates and maintains the lifecycle of the jumpbox and BOSH director.

It does not create any networks, security groups, or load balancers on vSphere.

## Create a Jumpbox and a BOSH Director

1. Export environment variables.
    ```
    export BBL_IAAS=vsphere
    export BBL_VSPHERE_VCENTER_USER
    export BBL_VSPHERE_VCENTER_PASSWORD
    export BBL_VSPHERE_VCENTER_IP
    export BBL_VSPHERE_VCENTER_DC
    export BBL_VSPHERE_VCENTER_CLUSTER
    export BBL_VSPHERE_VCENTER_RP
    export BBL_VSPHERE_NETWORK
    export BBL_VSPHERE_VCENTER_DS
    export BBL_VSPHERE_SUBNET_CIDR
    export BBL_VSPHERE_VCENTER_DISKS
    export BBL_VSPHERE_VCENTER_TEMPLATES
    export BBL_VSPHERE_VCENTER_VMS
    ```

    or powershell:

    ```powershell
    $env:BBL_IAAS="vsphere"
    $env:BBL_VSPHERE_VCENTER_USER=
    $env:BBL_VSPHERE_VCENTER_PASSWORD=
    $env:BBL_VSPHERE_VCENTER_IP=
    $env:BBL_VSPHERE_VCENTER_DC=
    $env:BBL_VSPHERE_VCENTER_CLUSTER=
    $env:BBL_VSPHERE_VCENTER_RP=
    $env:BBL_VSPHERE_NETWORK=
    $env:BBL_VSPHERE_VCENTER_DS=
    $env:BBL_VSPHERE_SUBNET_CIDR=
    $env:BBL_VSPHERE_VCENTER_DISKS=
    $env:BBL_VSPHERE_VCENTER_TEMPLATES=
    $env:BBL_VSPHERE_VCENTER_VMS=
    ```

1. Create jumpbox and bosh director.
    ```
    bbl up
    ```

## Next Steps

* [Target the BOSH Director](howto-target-bosh-director.md)