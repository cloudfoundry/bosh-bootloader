package application

import "github.com/pivotal-cf-experimental/bosh-bootloader/storage"

type GlobalConfiguration struct {
	Help               bool
	Version            bool
	EndpointOverride   string
	AWSAccessKeyID     string
	AWSSecretAccessKey string
	AWSRegion          string
	StateDir           string
}

type Configuration struct {
	Global          GlobalConfiguration
	Command         string
	SubcommandFlags []string
	State           storage.State
}
