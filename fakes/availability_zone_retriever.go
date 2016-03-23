package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type AvailabilityZoneRetriever struct {
	RetrieveCall struct {
		Returns struct {
			AZs   []string
			Error error
		}
	}
}

func (a *AvailabilityZoneRetriever) Retrieve(string, ec2.Client) ([]string, error) {
	return a.RetrieveCall.Returns.AZs, a.RetrieveCall.Returns.Error
}
