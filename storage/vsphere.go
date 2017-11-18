package storage

type VSphere struct {
	VCenterUser     string `json:"vcenterUser,omitempty"`
	VCenterPassword string `json:"vcenterPassword,omitempty"`
	VCenterIP       string `json:"vcenterIP,omitempty"`
	Datacenter      string `json:"datacenter,omitempty"`
	Cluster         string `json:"cluster,omitempty"`
	ResourcePool    string `json:"resourcePool,omitempty"`
	Network         string `json:"network,omitempty"`
	Datastore       string `json:"datastore,omitempty"`
	Subnet          string `json:"subnet,omitempty"`
}
