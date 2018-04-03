# Getting Started: Microsoft Azure

## Create a Service Principal Account

## bbl up

1. Export environment variables.
    ```
    export BBL_IAAS=azure
    export BBL_AZURE_CLIENT_ID=
    export BBL_AZURE_CLIENT_SECRET=
    export BBL_AZURE_REGION=
    export BBL_AZURE_SUBSCRIPTION_ID=
    export BBL_AZURE_TENANT_ID=
    ```
1. Create infrastructure, jumpbox, and bosh director.
    ```
    bbl up
    ```

## bbl up --lb-type cf --lb-cert cert.pfx --lb-key pfx-password

1. Follow the steps above.

1. To create cf load balancers for azure you must provide a certificate in the `.pfx` format:
    ```
    openssl genrsa -out DOMAIN_NAME.key 2048
    openssl req -new -x509 -days 365 -key DOMAIN_NAME.key -out DOMAIN_NAME.crt
    openssl pkcs12 -export -out PFX_FILE -inkey DOMAIN_NAME.key -in DOMAIN_NAME.crt
    ```

1. Save the password you entered when prompted by `openssl` to a file.
    ```
    echo SuperSecretPassword > PFX_FILE_PASSWORD
    ```
1. To `bbl  plan` or `bbl up` you can provide the pfx file and password:
    ```
    bbl plan --lb-type cf --lb-cert PFX_FILE --lb-key PFX_FILE_PASSWORD
    bbl up
    ```
