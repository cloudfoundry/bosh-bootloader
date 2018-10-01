package compute

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpcompute "google.golang.org/api/compute/v1"
)

type vpnTunnelsClient interface {
	ListVpnTunnels(region string) ([]*gcpcompute.VpnTunnel, error)
	DeleteVpnTunnel(region, vpnTunnel string) error
}

type VpnTunnels struct {
	client  vpnTunnelsClient
	logger  logger
	regions map[string]string
}

func NewVpnTunnels(client vpnTunnelsClient, logger logger, regions map[string]string) VpnTunnels {
	return VpnTunnels{
		client:  client,
		logger:  logger,
		regions: regions,
	}
}

func (v VpnTunnels) List(filter string) ([]common.Deletable, error) {
	tunnels := []*gcpcompute.VpnTunnel{}

	for _, region := range v.regions {
		l, err := v.client.ListVpnTunnels(region)
		if err != nil {
			return nil, fmt.Errorf("List Vpn Tunnels: %s", err)
		}

		tunnels = append(tunnels, l...)
	}

	var resources []common.Deletable

	for _, t := range tunnels {
		resource := NewVpnTunnel(v.client, t.Name, v.regions[t.Region])

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		proceed := v.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (VpnTunnels) Type() string {
	return "vpn-tunnel"
}
