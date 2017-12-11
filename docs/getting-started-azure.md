# Getting Started - Azure


## Azure Application Gateway

0. To create azure load balancers, prepare your certificates in the `.pfx` format:

```
openssl genrsa -out DOMAIN_NAME.key 2048
openssl req -new -x509 -days 365 -key DOMAIN_NAME.key -out DOMAIN_NAME.crt
openssl pkcs12 -export -out PFX_FILE -inkey DOMAIN_NAME.key -in DOMAIN_NAME.crt

# You will be asked to input the password for the ".pfx" certificate
```

0. To `bbl  plan` or `bbl up` you will provide that pfx file and password:

```
bbl up --lb-type cf --lb-cert PFX_FILE --lb-key PFX_FILE_PASSWORD
```
