package gcp

import compute "google.golang.org/api/compute/v1"

type Client interface {
	GetProject(projectID string) (*compute.Project, error)
	SetCommonInstanceMetadata(projectID string, metadata *compute.Metadata) (*compute.Operation, error)
}
