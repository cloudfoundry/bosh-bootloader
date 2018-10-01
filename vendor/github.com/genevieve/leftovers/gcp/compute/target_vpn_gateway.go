package compute

import "fmt"

type TargetVpnGateway struct {
	client targetVpnGatewaysClient
	name   string
	region string
}

func NewTargetVpnGateway(client targetVpnGatewaysClient, name, region string) TargetVpnGateway {
	return TargetVpnGateway{
		client: client,
		name:   name,
		region: region,
	}
}

func (t TargetVpnGateway) Delete() error {
	err := t.client.DeleteTargetVpnGateway(t.region, t.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (t TargetVpnGateway) Name() string {
	return t.name
}

func (TargetVpnGateway) Type() string {
	return "Target Vpn Gateway"
}
