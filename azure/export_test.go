package azure

func NewClientWithInjectedVMsClient(azureVMsClient AzureVMsClient) Client {
	return Client{
		azureVMsClient: azureVMsClient,
	}
}

func NewClientWithInjectedVNsClient(azureVNsClient AzureVNsClient) Client {
	return Client{
		azureVNsClient: azureVNsClient,
	}
}
