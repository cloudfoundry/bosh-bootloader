package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"     //nolint:staticcheck
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources" //nolint:staticcheck
	"golang.org/x/net/context"
)

type Client struct {
	azureVMsClient    AzureVMsClient
	azureGroupsClient AzureGroupsClient
}

type AzureVMsClient interface {
	List(ctx context.Context, resourceGroup string) (armcompute.VirtualMachineListResult, error)
}

type AzureGroupsClient interface {
	CheckExistence(ctx context.Context, resourceGroupName string, options *armresources.ResourceGroupsClientCheckExistenceOptions) (armresources.ResourceGroupsClientCheckExistenceResponse, error)
}

func (c Client) CheckExists(envID string) (bool, error) {
	resourceGroupName := fmt.Sprintf("%s-bosh", envID)

	response, err := c.azureGroupsClient.CheckExistence(context.TODO(), resourceGroupName, nil)
	if err != nil {
		return false, fmt.Errorf("Check existence for resource group %s: %s", resourceGroupName, err)
	}

	if response.Success == true {
		return true, nil
	}

	return false, nil
}

func (c Client) ValidateSafeToDelete(networkName string, envID string) error {
	resourceGroup := fmt.Sprintf("%s-bosh", envID)

	instances, err := c.azureVMsClient.List(context.TODO(), resourceGroup)
	if err != nil {
		return fmt.Errorf("List instances: %s", err)
	}

	for _, instance := range instances.Value {
		var vm string
		if instance.Name != nil {
			vm = *instance.Name
		}

		if instance.Tags == nil {
			return fmt.Errorf("bbl environment is not safe to delete; vms still exist in resource group: %s: %s", resourceGroup, vm)
		}

		tags := instance.Tags

		var deployment string
		if tags["deployment"] != nil {
			deployment = fmt.Sprintf(" (deployment: %s)", *tags["deployment"])
		}

		var job string
		if tags["job"] != nil {
			job = *tags["job"]
		}

		if job != "bosh" && job != "jumpbox" {
			return fmt.Errorf("bbl environment is not safe to delete; vms still exist in resource group: %s%s: %s", resourceGroup, deployment, vm)
		}
	}

	return nil
}
