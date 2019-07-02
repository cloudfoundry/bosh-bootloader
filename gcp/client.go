package gcp

import (
	"fmt"
	"strings"

	compute "google.golang.org/api/compute/v1"
)

type Client struct {
	computeClient ComputeClient
	projectID     string
	zone          string
}

type ComputeClient interface {
	ListInstances(projectID, zone string) (*compute.InstanceList, error)
	GetZones(region, projectID string) ([]string, error)
	GetZone(zone, projectID string) (*compute.Zone, error)
	GetRegion(region, projectID string) (*compute.Region, error)
	GetNetworks(name, projectID string) (*compute.NetworkList, error)
}

func (c Client) ProjectID() string {
	return c.projectID
}

func (c Client) listInstances() (*compute.InstanceList, error) {
	return c.computeClient.ListInstances(c.projectID, c.zone)
}

func (c Client) GetZones(region string) ([]string, error) {
	return c.computeClient.GetZones(region, c.projectID)
}

func (c Client) GetZone(zone string) (*compute.Zone, error) {
	return c.computeClient.GetZone(zone, c.projectID)
}

func (c Client) GetRegion(region string) (*compute.Region, error) {
	return c.computeClient.GetRegion(region, c.projectID)
}

func (c Client) GetNetworks(name string) (*compute.NetworkList, error) {
	return c.computeClient.GetNetworks(name, c.projectID)
}

// Methods added to conform to IAAS-agnostic interfaces

func (c Client) CheckExists(networkName string) (bool, error) {
	networkList, err := c.GetNetworks(networkName)
	if err != nil {
		return false, err
	}

	if len(networkList.Items) > 0 {
		return true, nil
	}
	return false, nil
}

func (c Client) ValidateSafeToDelete(networkName string, envID string) error {
	instanceList, err := c.listInstances()
	if err != nil {
		return err
	}

	var runningInstances []*compute.Instance
	for _, instance := range instanceList.Items {
		isInNetwork := c.isInNetwork(networkName, instance.NetworkInterfaces)
		isBoshDirector := c.isBoshDirector(instance.Metadata)

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

func (c Client) isInNetwork(networkName string, networkInterfaces []*compute.NetworkInterface) bool {
	for _, networkInterface := range networkInterfaces {
		if networkInterface.Network == networkName {
			return true
		}
	}

	return false
}

func (c Client) isBoshDirector(metadata *compute.Metadata) bool {
	for _, item := range metadata.Items {
		if item.Key == "director" && item.Value != nil && *item.Value == "bosh-init" {
			return true
		}
	}

	return false
}
