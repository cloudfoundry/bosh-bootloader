package storage

import "reflect"

type BOSH struct {
	DirectorUsername       string                 `json:"directorUsername"`
	DirectorPassword       string                 `json:"directorPassword"`
	DirectorAddress        string                 `json:"directorAddress"`
	DirectorSSLCA          string                 `json:"directorSSLCA"`
	DirectorSSLCertificate string                 `json:"directorSSLCertificate"`
	DirectorSSLPrivateKey  string                 `json:"directorSSLPrivateKey"`
	Credentials            map[string]string      `json:"credentials"`
	State                  map[string]interface{} `json:"state"`
	Manifest               string                 `json:"manifest"`
}

func (b BOSH) IsEmpty() bool {
	return reflect.DeepEqual(b, BOSH{})
}
