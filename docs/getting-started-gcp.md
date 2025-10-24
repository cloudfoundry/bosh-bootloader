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

## Cloud Config Features

### VM Extensions

The generated cloud-config for GCP includes several VM extensions that can be applied to your deployments:

- **`spot`**: Uses GCP Spot VMs for cost savings (~91% discount). Requires Google CPI 50.1.0 or later. Spot VMs may be preempted by GCP when capacity is needed.
- **`preemptible`**: Uses the legacy preemptible VM API (similar cost savings as spot, but spot is the recommended approach).
- **`internet-not-required`**: Disables external IPs for VMs that don't need internet access.
- **`100GB_ephemeral_disk`**, **`500GB_ephemeral_disk`**, **`1TB_ephemeral_disk`**: Provides larger ephemeral disks.

To use a VM extension in your deployment manifest:

```yaml
instance_groups:
- name: my-instance-group
  vm_extensions: [spot]
```

## Next Steps

* [Target the BOSH Director](howto-target-bosh-director.md)
* [Deploy Cloud Foundry](cloudfoundry.md)
* [Deploy Concourse](concourse.md)
