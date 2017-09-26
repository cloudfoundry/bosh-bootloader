package storage

type AWS struct {
	AccessKeyID     string `json:"accessKeyId,omitempty"`
	SecretAccessKey string `json:"secretAccessKey,omitempty"`
	Region          string `json:"region"`
}
