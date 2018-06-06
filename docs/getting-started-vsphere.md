# Getting Started: vSphere

`bbl` creates and maintains the lifecycle of the jumpbox and BOSH director.

It does not create any networks, security groups, or load balancers on vSphere.

## Jumbpox, Director

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
    export BBL_VSPHERE_SUBNET
    export BBL_VSPHERE_VCENTER_DISKS
    export BBL_VSPHERE_VCENTER_TEMPLATES
    export BBL_VSPHERE_VCENTER_VMS
    ```

1. Create jumpbox and bosh director.
    ```
    bbl up
    ```
