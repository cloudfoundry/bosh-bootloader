package storage

type VSphere struct {
	Cluster         string `json:"-"`
	Network         string `json:"-"`
	Subnet          string `json:"-"`
	VCenterUser     string `json:"-"`
	VCenterPassword string `json:"-"`
	VCenterIP       string `json:"-"`
	VCenterDC       string `json:"-"`
	VCenterRP       string `json:"-"`
	VCenterDS       string `json:"-"`
}
