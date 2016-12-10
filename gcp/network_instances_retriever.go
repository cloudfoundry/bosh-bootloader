package gcp

import "strings"

type NetworkInstancesRetriever struct {
	clientProvider clientProvider
}

func NewNetworkInstancesRetriever(clientProvider clientProvider, logger logger) NetworkInstancesRetriever {
	return NetworkInstancesRetriever{
		clientProvider: clientProvider,
	}
}

func (n NetworkInstancesRetriever) List(projectID, zone, networkName string) ([]string, error) {
	client := n.clientProvider.Client()
	instanceList, err := client.ListInstances(projectID, zone)
	if err != nil {
		return []string{}, err
	}

	instanceNames := []string{}
	for _, instance := range instanceList.Items {
		for _, networkInterface := range instance.NetworkInterfaces {
			if strings.Contains(networkInterface.Network, networkName) {
				instanceNames = append(instanceNames, instance.Name)
			}
		}
	}

	return instanceNames, nil
}
