package fakes

import compute "google.golang.org/api/compute/v1"

type GCPClient struct {
	ProjectIDCall struct {
		CallCount int
		Returns   struct {
			ProjectID string
		}
	}
	GetProjectCall struct {
		CallCount int
		Returns   struct {
			Project *compute.Project
			Error   error
		}
	}
	SetCommonInstanceMetadataCall struct {
		CallCount int
		Receives  struct {
			Metadata *compute.Metadata
		}
		Returns struct {
			Operation *compute.Operation
			Error     error
		}
	}
	ListInstancesCall struct {
		CallCount int
		Returns   struct {
			InstanceList *compute.InstanceList
			Error        error
		}
	}
	GetNetworksCall struct {
		CallCount int
		Receives  struct {
			Name string
		}
		Returns struct {
			NetworkList *compute.NetworkList
			Error       error
		}
	}
}

func (g *GCPClient) ProjectID() string {
	g.ProjectIDCall.CallCount++
	return g.ProjectIDCall.Returns.ProjectID
}

func (g *GCPClient) GetProject() (*compute.Project, error) {
	g.GetProjectCall.CallCount++
	return g.GetProjectCall.Returns.Project, g.GetProjectCall.Returns.Error
}

func (g *GCPClient) SetCommonInstanceMetadata(metadata *compute.Metadata) (*compute.Operation, error) {
	g.SetCommonInstanceMetadataCall.CallCount++
	g.SetCommonInstanceMetadataCall.Receives.Metadata = metadata
	return g.SetCommonInstanceMetadataCall.Returns.Operation, g.SetCommonInstanceMetadataCall.Returns.Error
}

func (g *GCPClient) ListInstances() (*compute.InstanceList, error) {
	g.ListInstancesCall.CallCount++
	return g.ListInstancesCall.Returns.InstanceList, g.ListInstancesCall.Returns.Error
}

func (g *GCPClient) GetNetworks(name string) (*compute.NetworkList, error) {
	g.GetNetworksCall.CallCount++
	g.GetNetworksCall.Receives.Name = name
	return g.GetNetworksCall.Returns.NetworkList, g.GetNetworksCall.Returns.Error
}
