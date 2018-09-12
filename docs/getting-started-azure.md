# Getting Started: Microsoft Azure

This guide is a walkthrough for deploying a BOSH director with `bbl`
on Microsoft Azure. Upon completion, you will have the following:

1. A BOSH director
1. A jumpbox
1. A set of randomly generated BOSH director credentials
1. A generated keypair allowing you to SSH into the BOSH director and
any instances BOSH deploys
1. A copy of the manifest the BOSH director was deployed with
1. A basic cloud config

## Create a Service Principal Account

You can use the cli utility [az-automation](https://github.com/genevieve/az-automation)
for creating a service principal account given you
have authenticated with the `az` cli.

The output will include your subscription id,
your tenant id, the client id, and the client secret.

These credentials will be passed to `bbl` so that
it can interact with Azure.

## Pave Infrastructure, Create a Jumpbox, and Create a BOSH Director

1. Export environment variables.
    ```
    export BBL_IAAS=azure
    export BBL_AZURE_CLIENT_ID=
    export BBL_AZURE_CLIENT_SECRET=
    export BBL_AZURE_REGION=
    export BBL_AZURE_SUBSCRIPTION_ID=
    export BBL_AZURE_TENANT_ID=
    ```

    or powershell:

    ```powershell
    $env:BBL_IAAS="azure"
    $env:BBL_AZURE_CLIENT_ID=
    $env:BBL_AZURE_CLIENT_SECRET=
    $env:BBL_AZURE_REGION=
    $env:BBL_AZURE_SUBSCRIPTION_ID=
    $env:BBL_AZURE_TENANT_ID=
    ```

1. Create infrastructure, jumpbox, and bosh director.
    ```
    bbl up
    ```

## Next Steps

* [Target the BOSH Director](howto-target-bosh-director.md)
* [Deploy Cloud Foundry](cloudfoundry.md)
* [Deploy Concourse](concourse.md)