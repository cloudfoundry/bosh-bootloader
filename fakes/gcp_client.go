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
	GetZonesCall struct {
		CallCount int
		Receives  struct {
			Region string
		}
		Returns struct {
			Zones []string
			Error error
		}
	}
	GetZoneCall struct {
		CallCount int
		Receives  struct {
			Zone string
		}
		Returns struct {
			Zone  *compute.Zone
			Error error
		}
	}
	GetRegionCall struct {
		CallCount int
		Receives  struct {
			Region string
		}
		Returns struct {
			Region *compute.Region
			Error  error
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

func (g *GCPClient) GetZones(region string) ([]string, error) {
	g.GetZonesCall.CallCount++
	g.GetZonesCall.Receives.Region = region
	return g.GetZonesCall.Returns.Zones, g.GetZonesCall.Returns.Error
}

func (g *GCPClient) GetZone(zone string) (*compute.Zone, error) {
	g.GetZoneCall.CallCount++
	g.GetZoneCall.Receives.Zone = zone
	return g.GetZoneCall.Returns.Zone, g.GetZoneCall.Returns.Error
}

func (g *GCPClient) GetRegion(region string) (*compute.Region, error) {
	g.GetRegionCall.CallCount++
	g.GetRegionCall.Receives.Region = region
	return g.GetRegionCall.Returns.Region, g.GetRegionCall.Returns.Error
}

func (g *GCPClient) GetNetworks(name string) (*compute.NetworkList, error) {
	g.GetNetworksCall.CallCount++
	g.GetNetworksCall.Receives.Name = name
	return g.GetNetworksCall.Returns.NetworkList, g.GetNetworksCall.Returns.Error
}
