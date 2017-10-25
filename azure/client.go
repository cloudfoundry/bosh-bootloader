package azure

import (
	"fmt"

	"github.com/Azure/azure-sdk-for-go/arm/compute"
	"github.com/Azure/azure-sdk-for-go/arm/network"
)

type Client struct {
	azureVMsClient AzureVMsClient
	azureVNsClient AzureVNsClient
}

type AzureVMsClient interface {
	List(resourceGroup string) (compute.VirtualMachineListResult, error)
}

type AzureVNsClient interface {
	List(networkName string) (network.VirtualNetworkListResult, error)
}

func (c Client) CheckExists(envID string) (bool, error) {
	resourceGroupName := fmt.Sprintf("%s-bosh", envID)
	networkName := fmt.Sprintf("%s-bosh-vn", envID)

	networkList, err := c.azureVNsClient.List(resourceGroupName)
	if err != nil {
		return false, fmt.Errorf("List networks: %s", err)
	}

	for _, network := range *networkList.Value {
		name := getOrEmpty(network.Name)
		if name == networkName {
			return true, nil
		}
	}

	return false, nil
}

func (c Client) ValidateSafeToDelete(networkName string, envID string) error {
	resourceGroupName := fmt.Sprintf("%s-bosh", envID)

	instanceList, err := c.azureVMsClient.List(resourceGroupName)
	if err != nil {
		return fmt.Errorf("List instances: %s", err)
	}

	for _, instance := range *instanceList.Value {
		vmName := getOrEmpty(instance.Name)
		if instance.Tags == nil {
			return fmt.Errorf("bbl environment is not safe to delete; vms still exist in resource group: %s: %s",
				resourceGroupName, vmName)
		}
		tags := *instance.Tags
		job := getOrEmpty(tags["job"])
		deployment := getOrEmpty(tags["deployment"])
		if deployment != "" {
			deployment = fmt.Sprintf(" (deployment: %s)", deployment)
		}
		if job != "bosh" && job != "jumpbox" {
			return fmt.Errorf("bbl environment is not safe to delete; vms still exist in resource group: %s%s: %s",
				resourceGroupName, deployment, vmName)
		}
	}

	return nil
}

func getOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
