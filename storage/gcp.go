package storage

type GCP struct {
	ServiceAccountKey string `json:"-"`
	// ideally we could get rid of ServiceAccountKeyPath,
	// but this is how we are passing the key along to terraform
	// ... we had trouble getting it to work with HEREDOC in go
	ServiceAccountKeyPath string   `json:"-"`
	ProjectID             string   `json:"-"`
	Zone                  string   `json:"zone"`
	Region                string   `json:"region"`
	Zones                 []string `json:"zones"`
}

func (g GCP) Empty() bool {
	return g.ServiceAccountKey == "" && g.ProjectID == "" && g.Region == "" && g.Zone == ""
}
