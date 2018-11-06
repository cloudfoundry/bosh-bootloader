package storage

type VSphere struct {
	Network          string `json:"-"`
	SubnetCIDR       string `json:"-"`
	VCenterCluster   string `json:"-"`
	VCenterUser      string `json:"-"`
	VCenterPassword  string `json:"-"`
	VCenterIP        string `json:"-"`
	VCenterDC        string `json:"-"`
	VCenterRP        string `json:"-"`
	VCenterDS        string `json:"-"`
	VCenterDisks     string `json:"-"`
	VCenterTemplates string `json:"-"`
	VCenterVMs       string `json:"-"`
}
