package storage

type State struct {
	Version        int       `json:"version"`
	BBLVersion     string    `json:"bblVersion"`
	IAAS           string    `json:"iaas"`
	ID             string    `json:"id"`
	EnvID          string    `json:"envID"`
	NoDirector     bool      `json:"noDirector"`
	AWS            AWS       `json:"aws,omitempty"`
	Azure          Azure     `json:"azure,omitempty"`
	GCP            GCP       `json:"gcp,omitempty"`
	VSphere        VSphere   `json:"vsphere,omitempty"`
	OpenStack      OpenStack `json:"openstack,omitempty"`
	Jumpbox        Jumpbox   `json:"jumpbox,omitempty"`
	BOSH           BOSH      `json:"bosh,omitempty"`
	TFState        string    `json:"tfState"`
	LB             LB        `json:"lb"`
	LatestTFOutput string    `json:"latestTFOutput"`
}
