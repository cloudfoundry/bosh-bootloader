# Testing

## Unit tests

```
ginkgo -r -p -race .
```

## Acceptance tests


### Google Cloud Platform

```
export BBL_GCP_SERVICE_ACCOUNT_KEY=
export LEFTOVERS_ACCEPTANCE=gcp

ginkgo -r -p -race acceptance
```


### Amazon Web Services

```
export BBL_AWS_ACCESS_KEY_ID=
export BBL_AWS_SECRET_ACCESS_KEY=
export BBL_AWS_REGION=
export LEFTOVERS_ACCEPTANCE=aws

ginkgo -r -p -race acceptance
```


### Microsoft Azure

```
export BBL_AZURE_SUBSCRIPTION_ID=
export BBL_AZURE_TENANT_ID=
export BBL_AZURE_CLIENT_ID=
export BBL_AZURE_CLIENT_SECRET=
export LEFTOVERS_ACCEPTANCE=azure

ginkgo -r -p -race acceptance
```

## vSphere

```
export BBL_VSPHERE_VCENTER_IP=
export BBL_VSPHERE_VCENTER_PASSWORD=
export BBL_VSPHERE_VCENTER_USER=
export BBL_VSPHERE_VCENTER_DC=
export LEFTOVERS_ACCEPTANCE=vsphere

ginkgo -r -p -race acceptance
```
