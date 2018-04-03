# Getting Started: vSphere

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
    ```

1. Create jumpbox and bosh director.
    ```
    bbl up
    ```
