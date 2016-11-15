package gcp

import compute "google.golang.org/api/compute/v1"

type Client interface {
	GetProject(projectID string) (*compute.Project, error)
	SetCommonInstanceMetadata(projectID string, metadata *compute.Metadata) (*compute.Operation, error)
}

type GCPClient struct {
	service *compute.Service
}

func (c GCPClient) GetProject(projectID string) (*compute.Project, error) {
	return c.service.Projects.Get(projectID).Do()
}

func (c GCPClient) SetCommonInstanceMetadata(projectID string, metadata *compute.Metadata) (*compute.Operation, error) {
	return c.service.Projects.SetCommonInstanceMetadata(projectID, metadata).Do()
}
