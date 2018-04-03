# Getting Started: GCP

## Create a Service Account

In order for `bbl` to interact with GCP, a service account must be created.

```
gcloud iam service-accounts create <service account name>

gcloud iam service-accounts keys create --iam-account='<service account name>@<project id>.iam.gserviceaccount.com' <service account name>.key.json

gcloud projects add-iam-policy-binding <project id> --member='serviceAccount:<service account name>@<project id>.iam.gserviceaccount.com' --role='roles/editor'
```

## Infrastructure, Jumpbox, Director

1. Export environment variables.
    ```
    export BBL_IAAS=gcp
    export BBL_GCP_REGION=
    export BBL_GCP_SERVICE_ACCOUNT_KEY=
    ```
1. Create infrastructure, jumpbox, and bosh director.
    ```
    bbl up
    ```

## + Cloud Foundry Load Balancers

To get all of the above plus load balancers for Cloud Foundry:

    ```
1. To `bbl  plan` or `bbl up` you can provide a cert, key, and (optionally) a domain:
    ```
    bbl plan --lb-type cf --lb-cert $CERT --lb-key $KEY --lb-comdin $DOMAIN
    bbl up
    ```
