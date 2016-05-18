package application

import "github.com/pivotal-cf-experimental/bosh-bootloader/storage"

type GlobalConfiguration struct {
	EndpointOverride string
	StateDir         string
}

type Configuration struct {
	Global          GlobalConfiguration
	Command         string
	SubcommandFlags []string
	State           storage.State
}
