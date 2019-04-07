# Getting Started: GCP

This guide is a walkthrough for deploying a BOSH director with `bbl`
on GCP. Upon completion, you will have the following:

1. A BOSH director
1. A jumpbox
1. A set of randomly generated BOSH director credentials
1. A generated keypair allowing you to SSH into the BOSH director and
any instances BOSH deploys
1. A copy of the manifest the BOSH director was deployed with
1. A basic cloud config

## Create a Service Account

In order for `bbl` to interact with GCP, a service account must be created.

```
gcloud iam service-accounts create <service account name>

gcloud iam service-accounts keys create --iam-account='<service account name>@<project id>.iam.gserviceaccount.com' <service account name>.key.json

gcloud projects add-iam-policy-binding <project id> --member='serviceAccount:<service account name>@<project id>.iam.gserviceaccount.com' --role='roles/editor'
```

### (Optional) Create a Custom GCP Cloud IAM Role for BBL

Optionally, a [custom IAM role](https://cloud.google.com/iam/docs/understanding-roles#custom_roles) can be created with a set of minimum permissions
rather than using the [primitive role](https://cloud.google.com/iam/docs/understanding-roles#primitive_role_definitions) definition of `roles/editor`.

The [bbl-iam-role.yml](bbl-iam-role.yml) can be used to create the `bbl` IAM role via the command line.
See [Creating a custom role using YAML files](https://cloud.google.com/iam/docs/creating-custom-roles#creating_a_custom_role).

Once created, the `add-iam-policy-binding` command above can now substitute `roles/editor` for the custom role definition: `projects/<project id>/roles/BBL`

## Pave Infrastructure, Create a Jumpbox, and Create a BOSH Director

1. Export environment variables.
    ```
    export BBL_IAAS=gcp
    export BBL_GCP_REGION=
    export BBL_GCP_SERVICE_ACCOUNT_KEY=
    ```

    or powershell:

    ```powershell
    $env:BBL_IAAS="gcp"
    $env:BBL_GCP_REGION=
    $env:BBL_GCP_SERVICE_ACCOUNT_KEY=
    ```
1. Create an empty directory to use as your bbl state directory.
    ```
    mkdir some-bbl-state-dir
    cd some-bbl-state-dir
    ```
1. Create infrastructure, jumpbox, and bosh director.
    ```
    bbl up
    ```

## Next Steps

* [Target the BOSH Director](howto-target-bosh-director.md)
* [Deploy Cloud Foundry](cloudfoundry.md)
* [Deploy Concourse](concourse.md)
