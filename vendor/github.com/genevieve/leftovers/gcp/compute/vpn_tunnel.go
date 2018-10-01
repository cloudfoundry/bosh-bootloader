package compute

import "fmt"

type VpnTunnel struct {
	client vpnTunnelsClient
	name   string
	region string
}

func NewVpnTunnel(client vpnTunnelsClient, name, region string) VpnTunnel {
	return VpnTunnel{
		client: client,
		name:   name,
		region: region,
	}
}

func (v VpnTunnel) Delete() error {
	err := v.client.DeleteVpnTunnel(v.region, v.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (v VpnTunnel) Name() string {
	return v.name
}

func (VpnTunnel) Type() string {
	return "Vpn Tunnel"
}
