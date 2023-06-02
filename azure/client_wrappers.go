package azure

import (
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute" //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage" //nolint:staticcheck
	"golang.org/x/net/context"
)

type AzureVMsClientWrapper struct {
	Client *armcompute.VirtualMachinesClient
}

func (c AzureVMsClientWrapper) List(ctx context.Context, resourceGroup string) (armcompute.VirtualMachineListResult, error) {
	pager := c.Client.NewListAllPager(nil)
	result := make([]*armcompute.VirtualMachine, 0)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}
		result = append(result, nextResult.Value...)
	}

	return armcompute.VirtualMachineListResult{Value: result}, nil
}

type AzureStorageClientWrapper struct {
	Client *armstorage.AccountsClient
}

func (c AzureStorageClientWrapper) List(ctx context.Context) (armstorage.AccountListResult, error) {
	pager := c.Client.NewListPager(nil)
	result := make([]*armstorage.Account, 0)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to advance page: %v", err)
		}
		result = append(result, nextResult.Value...)
	}

	return armstorage.AccountListResult{Value: result}, nil
}

