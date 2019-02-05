package storage

type OpenStack struct {
	AuthURL        string   `json:"-"`
	AZ             string   `json:"-"`
	NetworkID      string   `json:"-"`
	NetworkName    string   `json:"-"`
	Password       string   `json:"-"`
	Username       string   `json:"-"`
	Project        string   `json:"-"`
	Domain         string   `json:"-"`
	Region         string   `json:"-"`
	CACertFile     string   `json:"-"`
	Insecure       string   `json:"-"`
	DNSNameServers []string `json:"-"`
}
