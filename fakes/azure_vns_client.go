package fakes

import "github.com/Azure/azure-sdk-for-go/arm/network"

type AzureVNsClient struct {
	ListCall struct {
		CallCount int
		Receives  struct {
			ResourceGroup string
		}
		Returns struct {
			Result network.VirtualNetworkListResult
			Error  error
		}
	}
}

func (a *AzureVNsClient) List(resourceGroup string) (network.VirtualNetworkListResult, error) {
	a.ListCall.CallCount++
	a.ListCall.Receives.ResourceGroup = resourceGroup
	return a.ListCall.Returns.Result, a.ListCall.Returns.Error
}
