# Generic Steps for Cloud Foundry Deployment

1. You can use existing certificate and key files, or generate new ones. See below for instructions on generating these files for Microsoft Azure.
    `bbl plan \ 
          --lb-type cf \
          --lb-cert <PATH_TO_CERT_FILE> \
          --lb-key <PATH_TO_KEY_FILE> \
          --lb-domain <DOMAIN_NAME> \
          && bbl up`.

1. Create an environment and target the BOSH director with `eval "$(bbl print-env)"`

1. `bosh deploy cf-deployment.yml -o operations/<MY IaaS> -v system_domain=<DOMAIN_NAME>` using the [CF deployment manifest!](https://github.com/cloudfoundry/cf-deployment)

## Appendix: Generating Load Balancer Key and Certificate Files for Microsoft Azure

To create Cloud Foundry load balancers for Microsoft Azure you must provide a certificate
in the `.pfx` format:

```
openssl genrsa -out DOMAIN_NAME.key 2048
openssl req -new -x509 -days 365 -key DOMAIN_NAME.key -out DOMAIN_NAME.crt
openssl pkcs12 -export -out PFX_FILE -inkey DOMAIN_NAME.key -in DOMAIN_NAME.crt
```

Save the password you entered when prompted by `openssl` to a file.

```
echo SuperSecretPassword > PFX_FILE_PASSWORD
```

To `bbl  plan` or `bbl up` you can provide the `.pfx` file and password:

```
bbl plan --lb-type cf --lb-cert PFX_FILE --lb-key PFX_FILE_PASSWORD
```
