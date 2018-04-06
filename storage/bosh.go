package storage

import "reflect"

type BOSH struct {
	DirectorName           string                 `json:"directorName"`
	DirectorUsername       string                 `json:"directorUsername"`
	DirectorPassword       string                 `json:"directorPassword"`
	DirectorAddress        string                 `json:"directorAddress"`
	DirectorSSLCA          string                 `json:"directorSSLCA"`
	DirectorSSLCertificate string                 `json:"directorSSLCertificate"`
	DirectorSSLPrivateKey  string                 `json:"directorSSLPrivateKey"`
	Variables              string                 `json:"variables,omitempty"`
	State                  map[string]interface{} `json:"state,omitempty"`
	Manifest               string                 `json:"manifest,omitempty"`
}

func (b BOSH) IsEmpty() bool {
	return reflect.DeepEqual(b, BOSH{})
}
