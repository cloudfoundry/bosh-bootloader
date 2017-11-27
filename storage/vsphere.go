package storage

type VSphere struct {
	Cluster         string `json:"cluster,omitempty"`
	Network         string `json:"network,omitempty"`
	Subnet          string `json:"subnet,omitempty"`
	VCenterUser     string `json:"-"`
	VCenterPassword string `json:"-"`
	VCenterIP       string `json:"vcenterIP,omitempty"`
	VCenterDC       string `json:"vcenterDC,omitempty"`
	VCenterRP       string `json:"vcenterRP,omitempty"`
	VCenterDS       string `json:"vcenterDS,omitempty"`
}
