package storage

type CloudStack struct {
	Network            string `json:"-"`
	Subnet             string `json:"-"`
	Endpoint           string `json:"-"`
	ApiKey             string `json:"-"`
	SecretAccessKey    string `json:"-"`
	Zone               string `json:"-"`
	PrivateKey         string `json:"-"`
	InternalCidr       string `json:"-"`
	ExternalIP         string `json:"-"`
	NetworkVpcOffering string `json:"-"`
	ComputeOffering    string `json:"-"`
	IsoSegment         bool   `json:"-"`
}
