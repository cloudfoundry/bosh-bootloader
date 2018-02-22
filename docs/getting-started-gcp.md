# Getting started gcp

## Creating a service account

In order for `bbl` to interact with GCP, a service account must be created.

```
gcloud iam service-accounts create <service account name>

gcloud iam service-accounts keys create --iam-account='<service account name>@<project id>.iam.gserviceaccount.com' <service account name>.key.json

gcloud projects add-iam-policy-binding <project id> --member='serviceAccount:<service account name>@<project id>.iam.gserviceaccount.com' --role='roles/editor'
```

