package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/application"

type CommandLineParser struct {
	ParseCall struct {
		Receives struct {
			Arguments []string
		}
		Returns struct {
			CommandLineConfiguration application.CommandLineConfiguration
			Error                    error
		}
	}
}

func (p *CommandLineParser) Parse(arguments []string) (application.CommandLineConfiguration, error) {
	p.ParseCall.Receives.Arguments = arguments
	return p.ParseCall.Returns.CommandLineConfiguration, p.ParseCall.Returns.Error
}
