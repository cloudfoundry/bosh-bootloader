# Generic Steps for Cloud Foundry Deployment

1. Create an environment and target the BOSH director with `eval "$(bbl print-env)"`

1. `bbl plan --lb-type cf --lb-cert <PATH_TO_CERT_FILE> --lb-key <PATH_TO_KEY_FILE> && bbl up`. You can use existing certificate and key files, or generate new ones. See below for instructions on generating these files for Microsoft Azure.

1. `bosh deploy cf-deployment.yml -o operations/<MY IaaS>` using the [CF deployment manifest!](https://github.com/cloudfoundry/cf-deployment)

## Appendix: Generating Load Balancer Key and Certificate Files for Microsoft Azure

To create Cloud Foundry load balancers for Microsoft Azure you must provide a certificate
in the `.pfx` format:

```
export DOMAIN="<INSERT-DOMAIN-NAME-HERE>"
openssl genrsa -out $DOMAIN.key 2048
openssl req -new -x509 -days 365 -key $DOMAIN.key -out $DOMAIN.crt
openssl pkcs12 -export -out PFX_FILE -inkey $DOMAIN.key -in $DOMAIN.crt
```

Save the password you entered when prompted by `openssl` to a file.

```
export PFX_FILE_PASSWORD="<INSERT-FILE-PATH-HERE>"
echo SuperSecretPassword > $PFX_FILE_PASSWORD
```

To `bbl  plan` or `bbl up` you can provide the `.pfx` file and password:

```
export PFX_FILE="<INSERT-FILE-PATH-HERE>"
bbl plan --lb-type cf --lb-cert $PFX_FILE --lb-key $PFX_FILE_PASSWORD
```
