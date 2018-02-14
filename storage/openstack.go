package storage

type OpenStack struct {
	InternalCidr         string `json:"-"`
	ExternalIP           string `json:"-"`
	AuthURL              string `json:"-"`
	AZ                   string `json:"-"`
	DefaultKeyName       string `json:"-"`
	DefaultSecurityGroup string `json:"-"`
	NetworkID            string `json:"-"`
	Password             string `json:"-"`
	Username             string `json:"-"`
	Project              string `json:"-"`
	Domain               string `json:"-"`
	Region               string `json:"-"`
	PrivateKey           string `json:"-"`
}
