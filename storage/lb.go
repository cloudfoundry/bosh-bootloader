package storage

type LB struct {
	Type   string `json:"type"`
	Cert   string `json:"cert"`
	Key    string `json:"key"`
	Domain string `json:"domain,omitempty"`
}
