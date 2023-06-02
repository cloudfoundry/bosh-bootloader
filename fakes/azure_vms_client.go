package fakes

import (
	"context"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute" //nolint:staticcheck
)

type AzureVMsClient struct {
	ListCall struct {
		CallCount int
		Receives  struct {
			ResourceGroup string
		}
		Returns struct {
			Result armcompute.VirtualMachineListResult
			Error  error
		}
	}
}

func (a *AzureVMsClient) List(ctx context.Context, resourceGroup string) (armcompute.VirtualMachineListResult, error) {
	a.ListCall.CallCount++
	a.ListCall.Receives.ResourceGroup = resourceGroup
	return a.ListCall.Returns.Result, a.ListCall.Returns.Error
}
