package storage

type GCP struct {
	ServiceAccountKey string   `json:"-"`
	ProjectID         string   `json:"-"`
	Zone              string   `json:"zone"`
	Region            string   `json:"region"`
	Zones             []string `json:"zones"`
}

func (g GCP) Empty() bool {
	return g.ServiceAccountKey == "" && g.ProjectID == "" && g.Region == "" && g.Zone == ""
}
