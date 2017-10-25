package azure

func NewClientWithInjectedVMsClient(azureVMsClient AzureVMsClient) Client {
	return Client{
		azureVMsClient: azureVMsClient,
	}
}

func NewClientWithInjectedGroupsClient(azureGroupsClient AzureGroupsClient) Client {
	return Client{
		azureGroupsClient: azureGroupsClient,
	}
}
