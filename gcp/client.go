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

func (c GCPClient) GetNetworks(name string) (*compute.NetworkList, error) {
	networksListCall := c.service.Networks.List(c.projectID)
	return networksListCall.Filter(fmt.Sprintf("name eq %s", name)).Do()
}
