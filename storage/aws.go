package storage

type AWS struct {
	AccessKeyID     string `json:"-"`
	SecretAccessKey string `json:"-"`
	AssumeRoleArn   string `json:"assumeRole,omitempty"`
	Region          string `json:"region,omitempty"`
}
