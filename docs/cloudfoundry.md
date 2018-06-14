## Generic Steps for Cloud Foundry Deployment

1. Create an environment and target the BOSH director with `bbl print-env | eval`

1. `bbl plan --lb-type cf --lb-cert cert --lb-key key && bbl up` with a certificate and key as flags or environment variables.
(Continue to provide the IaaS credentials as flags or environment variables.)

1. `bosh deploy cf-deployment.yml -o operations/<MY IaaS>` using the [CF deployment manifest!](https://github.com/cloudfoundry/cf-deployment)
