package fakes

import compute "google.golang.org/api/compute/v1"

type GCPProjectsService struct {
	GetCall struct {
		CallCount int
		Receives  struct {
			ProjectID string
		}
		Returns struct {
			Project *compute.Project
			Error   error
		}
	}
	SetCommonInstanceMetadataCall struct {
		CallCount int
		Receives  struct {
			ProjectID string
			Metadata  *compute.Metadata
		}
		Returns struct {
			Operation *compute.Operation
			Error     error
		}
	}
}

func (g *GCPProjectsService) Get(projectID string) (*compute.Project, error) {
	g.GetCall.CallCount++
	g.GetCall.Receives.ProjectID = projectID
	return g.GetCall.Returns.Project, g.GetCall.Returns.Error
}

func (g *GCPProjectsService) SetCommonInstanceMetadata(projectID string, metadata *compute.Metadata) (*compute.Operation, error) {
	g.SetCommonInstanceMetadataCall.CallCount++
	g.SetCommonInstanceMetadataCall.Receives.ProjectID = projectID
	g.SetCommonInstanceMetadataCall.Receives.Metadata = metadata
	return g.SetCommonInstanceMetadataCall.Returns.Operation, g.SetCommonInstanceMetadataCall.Returns.Error
}
