package gcp

import (
	"errors"
	"strings"

	compute "google.golang.org/api/compute/v1"
)

type NetworkInstancesChecker struct {
	clientProvider clientProvider
}

func NewNetworkInstancesChecker(clientProvider clientProvider) NetworkInstancesChecker {
	return NetworkInstancesChecker{
		clientProvider: clientProvider,
	}
}

func (n NetworkInstancesChecker) ValidateSafeToDelete(networkName string) error {
	client := n.clientProvider.Client()
	instanceList, err := client.ListInstances()
	if err != nil {
		return err
	}

	for _, instance := range instanceList.Items {
		isInNetwork := n.isInNetwork(networkName, instance.NetworkInterfaces)
		isBoshDirector := n.isBoshDirector(instance.Metadata)

		if isInNetwork && !isBoshDirector {
			return errors.New("bbl environment is not safe to delete; vms still exist in network")
		}
	}

	return nil
}

func (n NetworkInstancesChecker) isInNetwork(networkName string, networkInterfaces []*compute.NetworkInterface) bool {
	for _, networkInterface := range networkInterfaces {
		return strings.Contains(networkInterface.Network, networkName)
	}

	return false
}

func (n NetworkInstancesChecker) isBoshDirector(metadata *compute.Metadata) bool {
	for _, metadata := range metadata.Items {
		return metadata.Key == "director" && metadata.Value != nil && *metadata.Value == "bosh-init"
	}

	return false
}
