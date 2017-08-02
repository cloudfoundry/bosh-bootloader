package gcp

import compute "google.golang.org/api/compute/v1"

type metadataSetter interface {
	GetProject() (*compute.Project, error)
	SetCommonInstanceMetadata(metadata *compute.Metadata) (*compute.Operation, error)
}

type instanceLister interface {
	ListInstances() (*compute.InstanceList, error)
}

type logger interface {
	Step(string, ...interface{})
}
