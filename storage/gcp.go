package storage

type GCP struct {
	ServiceAccountKey     string   `json:"-"`
	ServiceAccountKeyPath string   `json:"-"`
	ProjectID             string   `json:"-"`
	Zone                  string   `json:"zone,omitempty"`
	Region                string   `json:"region,omitempty"`
	Zones                 []string `json:"zones,omitempty"`
}

func (g GCP) Empty() bool {
	return g.ServiceAccountKey == "" && g.ProjectID == "" && g.Region == "" && g.Zone == ""
}
