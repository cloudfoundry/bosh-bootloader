# Leftovers :turkey:

Go cli & library for cleaning up **orphaned IaaS resources**.

* <a href='#why'>Why might you use this?</a>
* <a href='#what'>What's it look like?</a>
* <a href='#how'>Installation</a>
* <a href='#usage'>Usage</a>
* [Resources you can delete by IaaS.](RESOURCES.md)



## <a name='why'></a> Why might you use this?
- You `terraform apply`'d way back when and lost your `terraform.tfstate`
- You used the console or cli to create some infrastructure and want to clean up
- Your acceptance tests in CI failed, the container disappeared, and
infrastructure resources were tragically orphaned



## <a name='what'></a>What's it look like?

It will **prompt you before deleting** any resource by default, ie:

```css
> leftovers --filter banana

[Firewall: banana-http] Delete? (y/N)
```

It can be configured to **not** prompt, ie:

```css
> leftovers --filter banana --no-confirm

[Firewall: banana-http] Deleting...
[Firewall: banana-http] Deleted!
```

Or maybe you want to **see all of the resources** in your IaaS, ie:
```css
> leftovers --filter banana --dry-run

[Firewall: banana-http]
[Network: banana]
```


Finally, you might want to delete a single resource type::
```css
> leftovers types
service-account

> leftovers --filter banana --type service-account --no-confirm
[Service Account: banana@pivotal.io] Deleting...
[Service Account: banana@pivotal.io] Deleted!
```



## <a name='how'></a>Installation

### Option 1
[Install go.](https://golang.org/doc/install) Then:

```
go get -u github.com/genevieve/leftovers/cmd/leftovers
```

### Option 2

```
brew tap genevieve/tap
brew install leftovers
```



## <a name='how'></a>Usage

```
Usage:
  leftovers [OPTIONS]

Application Options:
  -i, --iaas=                     The IaaS for clean up. (default: aws) [$BBL_IAAS]
  -n, --no-confirm                Destroy resources without prompting. This is dangerous, make good choices!
  -f, --filter=                   Filtering resources by an environment name.
  -d, --dry-run                   List all resources without deleting any.
      --aws-access-key-id=        AWS access key id. [$BBL_AWS_ACCESS_KEY_ID]
      --aws-secret-access-key=    AWS secret access key. [$BBL_AWS_SECRET_ACCESS_KEY]
      --aws-region=               AWS region. [$BBL_AWS_REGION]
      --azure-client-id=          Azure client id. [$BBL_AZURE_CLIENT_ID]
      --azure-client-secret=      Azure client secret. [$BBL_AZURE_CLIENT_SECRET]
      --azure-tenant-id=          Azure tenant id. [$BBL_AZURE_TENANT_ID]
      --azure-subscription-id=    Azure subscription id. [$BBL_AZURE_SUBSCRIPTION_ID]
      --gcp-service-account-key=  GCP service account key path. [$BBL_GCP_SERVICE_ACCOUNT_KEY]
      --vsphere-vcenter-ip=       vSphere vCenter IP address. [$BBL_VSPHERE_VCENTER_IP]
      --vsphere-vcenter-password= vSphere vCenter password. [$BBL_VSPHERE_VCENTER_PASSWORD]
      --vsphere-vcenter-user=     vSphere vCenter username. [$BBL_VSPHERE_VCENTER_USER]
      --vsphere-vcenter-dc=       vSphere vCenter datacenter. [$BBL_VSPHERE_VCENTER_DC]

Help Options:
  -h, --help                     Show this help message
```

