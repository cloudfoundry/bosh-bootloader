package gcp

import (
	"fmt"

	compute "google.golang.org/api/compute/v1"
)

type Client interface {
	ProjectID() string
	GetProject() (*compute.Project, error)
	SetCommonInstanceMetadata(metadata *compute.Metadata) (*compute.Operation, error)
	ListInstances() (*compute.InstanceList, error)
	GetZones(region string) (*compute.ZoneList, error)
	GetZone(zone string) (*compute.Zone, error)
	GetRegion(region string) (*compute.Region, error)
	GetNetworks(name string) (*compute.NetworkList, error)
}

type GCPClient struct {
	service   *compute.Service
	projectID string
	zone      string
}

func (c GCPClient) ProjectID() string {
	return c.projectID
}

func (c GCPClient) GetProject() (*compute.Project, error) {
	return c.service.Projects.Get(c.projectID).Do()
}

func (c GCPClient) SetCommonInstanceMetadata(metadata *compute.Metadata) (*compute.Operation, error) {
	return c.service.Projects.SetCommonInstanceMetadata(c.projectID, metadata).Do()
}

func (c GCPClient) ListInstances() (*compute.InstanceList, error) {
	return c.service.Instances.List(c.projectID, c.zone).Do()
}

func (c GCPClient) GetZones(region string) (*compute.ZoneList, error) {
	zonesListCall := c.service.Zones.List(c.projectID)
	return zonesListCall.Filter(fmt.Sprintf("region eq %s", region)).Do()
}

func (c GCPClient) GetZone(zone string) (*compute.Zone, error) {
	return c.service.Zones.Get(c.projectID, zone).Do()
}

func (c GCPClient) GetRegion(region string) (*compute.Region, error) {
	return c.service.Regions.Get(c.projectID, region).Do()
}

func (c GCPClient) GetNetworks(name string) (*compute.NetworkList, error) {
	networksListCall := c.service.Networks.List(c.projectID)
	return networksListCall.Filter(fmt.Sprintf("name eq %s", name)).Do()
}
