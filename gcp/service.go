package gcp

import compute "google.golang.org/api/compute/v1"

type ProjectsService interface {
	Get(project string) (*compute.Project, error)
	SetCommonInstanceMetadata(project string, metadata *compute.Metadata) (*compute.Operation, error)
}

type ServiceWrapper interface {
	GetProjectsService() ProjectsService
}
