package gcp

import (
	"fmt"
	"strings"

	compute "google.golang.org/api/compute/v1"
)

type NetworkInstancesChecker struct {
	client instanceLister
}

func NewNetworkInstancesChecker(client instanceLister) NetworkInstancesChecker {
	return NetworkInstancesChecker{
		client: client,
	}
}

func (n NetworkInstancesChecker) ValidateSafeToDelete(networkName string) error {
	instanceList, err := n.client.ListInstances()
	if err != nil {
		return err
	}

	var runningInstances []*compute.Instance
	for _, instance := range instanceList.Items {
		isInNetwork := n.isInNetwork(networkName, instance.NetworkInterfaces)
		isBoshDirector := n.isBoshDirector(instance.Metadata)

		if isInNetwork && !isBoshDirector {
			runningInstances = append(runningInstances, instance)
		}
	}

	if len(runningInstances) == 0 {
		return nil
	}

	var errorMessages []string
	for _, instance := range runningInstances {
		var hasDeployment bool
		for _, item := range instance.Metadata.Items {
			if item.Key == "deployment" {
				errorMessages = append(errorMessages, fmt.Sprintf("%s (deployment: %s)", instance.Name, *item.Value))
				hasDeployment = true
				break
			}
		}

		if !hasDeployment {
			errorMessages = append(errorMessages, fmt.Sprintf("%s (not managed by bosh)", instance.Name))
		}
	}

	return fmt.Errorf("bbl environment is not safe to delete; vms still exist in network:\n%s",
		strings.Join(errorMessages, "\n"))
}

func (n NetworkInstancesChecker) isInNetwork(networkName string, networkInterfaces []*compute.NetworkInterface) bool {
	for _, networkInterface := range networkInterfaces {
		if strings.Contains(networkInterface.Network, networkName) {
			return true
		}
	}

	return false
}

func (n NetworkInstancesChecker) isBoshDirector(metadata *compute.Metadata) bool {
	for _, item := range metadata.Items {
		if item.Key == "director" && item.Value != nil && *item.Value == "bosh-init" {
			return true
		}
	}

	return false
}
