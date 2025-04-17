package gcp

import (
	"fmt"
	"sort"

	"google.golang.org/api/compute/v1"
)

type gcpComputeClient struct {
	service *compute.Service
}

func (g gcpComputeClient) ListInstances(projectID, zone string) (*compute.InstanceList, error) {
	return g.service.Instances.List(projectID, zone).Do()
}

func (g gcpComputeClient) GetZones(region, projectID string) ([]string, error) {
	regionCall, err := g.GetRegion(region, projectID)
	if err != nil {
		return []string{}, err
	}

	zoneURLs := map[string]struct{}{}
	for _, zoneURL := range regionCall.Zones {
		zoneURLs[zoneURL] = struct{}{}
	}

	zonesInProject, err := g.service.Zones.List(projectID).Do()
	if err != nil {
		return []string{}, err
	}

	zonesInRegion := []string{}
	for _, zone := range zonesInProject.Items {
		if _, ok := zoneURLs[zone.SelfLink]; ok {
			zonesInRegion = append(zonesInRegion, zone.Name)
		}
	}

	sort.Strings(zonesInRegion)
	return zonesInRegion, nil
}

func (g gcpComputeClient) GetZone(zone, projectID string) (*compute.Zone, error) {
	return g.service.Zones.Get(projectID, zone).Do()
}

func (g gcpComputeClient) GetRegion(region, projectID string) (*compute.Region, error) {
	return g.service.Regions.Get(projectID, region).Do()
}

func (g gcpComputeClient) GetNetworks(name, projectID string) (*compute.NetworkList, error) {
	networksListCall := g.service.Networks.List(projectID)
	return networksListCall.Filter(fmt.Sprintf("name eq %s", name)).Do()
}
