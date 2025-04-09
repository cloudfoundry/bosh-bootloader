package fakes

import "google.golang.org/api/compute/v1"

type GCPComputeClient struct {
	ListInstancesCall struct {
		CallCount int
		Receives  struct {
			ProjectID string
			Zone      string
		}
		Returns struct {
			InstanceList *compute.InstanceList
			Error        error
		}
	}
	GetZonesCall struct {
		CallCount int
		Receives  struct {
			Region    string
			ProjectID string
		}
		Returns struct {
			Zones []string
			Error error
		}
	}
	GetZoneCall struct {
		CallCount int
		Receives  struct {
			Zone      string
			ProjectID string
		}
		Returns struct {
			Zone  *compute.Zone
			Error error
		}
	}
	GetRegionCall struct {
		CallCount int
		Receives  struct {
			Region    string
			ProjectID string
		}
		Returns struct {
			Region *compute.Region
			Error  error
		}
	}
	GetNetworksCall struct {
		CallCount int
		Receives  struct {
			Name      string
			ProjectID string
		}
		Returns struct {
			NetworkList *compute.NetworkList
			Error       error
		}
	}
}

func (g *GCPComputeClient) ListInstances(projectID, zone string) (*compute.InstanceList, error) {
	g.ListInstancesCall.CallCount++
	g.ListInstancesCall.Receives.ProjectID = projectID
	g.ListInstancesCall.Receives.Zone = zone
	return g.ListInstancesCall.Returns.InstanceList, g.ListInstancesCall.Returns.Error
}

func (g *GCPComputeClient) GetZones(region, projectID string) ([]string, error) {
	g.GetZonesCall.CallCount++
	g.GetZonesCall.Receives.Region = region
	g.GetZonesCall.Receives.ProjectID = projectID
	return g.GetZonesCall.Returns.Zones, g.GetZonesCall.Returns.Error
}

func (g *GCPComputeClient) GetZone(zone, projectID string) (*compute.Zone, error) {
	g.GetZoneCall.CallCount++
	g.GetZoneCall.Receives.Zone = zone
	g.GetZoneCall.Receives.ProjectID = projectID
	return g.GetZoneCall.Returns.Zone, g.GetZoneCall.Returns.Error
}

func (g *GCPComputeClient) GetRegion(region, projectID string) (*compute.Region, error) {
	g.GetRegionCall.CallCount++
	g.GetRegionCall.Receives.Region = region
	g.GetRegionCall.Receives.ProjectID = projectID
	return g.GetRegionCall.Returns.Region, g.GetRegionCall.Returns.Error
}

func (g *GCPComputeClient) GetNetworks(name, projectID string) (*compute.NetworkList, error) {
	g.GetNetworksCall.CallCount++
	g.GetNetworksCall.Receives.Name = name
	g.GetNetworksCall.Receives.ProjectID = projectID
	return g.GetNetworksCall.Returns.NetworkList, g.GetNetworksCall.Returns.Error
}
