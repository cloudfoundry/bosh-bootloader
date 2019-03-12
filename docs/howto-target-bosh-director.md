# How To Target The BOSH Director

## The easy way: `bbl print-env`

**COMMON GOTCHA:** The quotes are necessary to successfully target the director!
```
eval "$(bbl print-env)"
```

or powershell:

```powershell
iex $(bbl print-env | Out-String)
```
## Alternatives to `bbl print-env`

Separate commands are available for the `bbl print-env` fields:

```
$ bbl director-address
https://10.0.0.6:25555

$ bbl director-username
user-d3783rk

$ bbl director-password
p-23dah71skl

$ bbl director-ca-cert
-----BEGIN CERTIFICATE-----
MIIDtzCCAp+gAwIBAgIJAIPgaUgWRCE8MA0GCSqGSIb3DQEBBQUAMEUxCzAJBgNV
...
-----END CERTIFICATE-----
```

You might save the CA certificate to a file:

```
$ bbl director-ca-cert > bosh.crt
$ export BOSH_CA_CERT=bosh.crt
```

```powershell
bbl director-ca-cert | Out-File bosh.crt
$env:BOSH_CA_CERT="bosh.crt"
```

To login:

```
$ export BOSH_ENVIRONMENT=$(bbl director-address)
$ bosh alias-env <INSERT TARGET NAME>
$ bosh log-in
Username: user-d3783rk
Password: p-23dah71sk1
```

or powershell:

```powershell
$env:TARGET_NAME="example-name" # Assigns a name to the created environment for easier access.
$env:BOSH_ENVIRONMENT=$(bbl director-address)
bosh alias-env $env:TARGET_NAME
bosh log-in
```

Now you're ready to deploy software with BOSH.
