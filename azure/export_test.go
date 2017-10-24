package azure

func NewClientWithInjectedVMsClient(azureVMsClient AzureVMsClient) Client {
	return Client{
		azureVMsClient: azureVMsClient,
	}
}
