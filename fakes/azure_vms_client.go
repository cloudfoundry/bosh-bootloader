package fakes

import (
	"github.com/Azure/azure-sdk-for-go/arm/compute" //nolint:staticcheck
)

type AzureVMsClient struct {
	ListCall struct {
		CallCount int
		Receives  struct {
			ResourceGroup string
		}
		Returns struct {
			Result compute.VirtualMachineListResult
			Error  error
		}
	}
}

func (a *AzureVMsClient) List(resourceGroup string) (compute.VirtualMachineListResult, error) {
	a.ListCall.CallCount++
	a.ListCall.Receives.ResourceGroup = resourceGroup
	return a.ListCall.Returns.Result, a.ListCall.Returns.Error
}
