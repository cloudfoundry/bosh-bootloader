package actors

type TFState struct {
	Modules []Module
}

type Module struct {
	Resources map[string]Resource
}

type Resource struct {
	Primary Primary
}

type Primary struct {
	Attributes map[string]string
}

func (tfState TFState) GetNetworkName() string {
	return tfState.Modules[0].Resources["google_compute_network.bbl"].Primary.Attributes["name"]
}

func (tfState TFState) GetSubnetName() string {
	return tfState.Modules[0].Resources["google_compute_subnetwork.bbl-subnet"].Primary.Attributes["name"]
}
