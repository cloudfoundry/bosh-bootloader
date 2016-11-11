package fakes

import "google.golang.org/api/compute/v1"

type GCPProjectAccessor interface {
	SetCommonInstanceMetadata(project string, metadata *compute.Metadata) *compute.ProjectsSetCommonInstanceMetadataCall
}
