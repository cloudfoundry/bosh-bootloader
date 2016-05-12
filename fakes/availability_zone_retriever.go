package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type AvailabilityZoneRetriever struct {
	RetrieveCall struct {
		Receives struct {
			Region    string
			EC2Client ec2.Client
		}
		Returns struct {
			AZs   []string
			Error error
		}
	}
}

func (a *AvailabilityZoneRetriever) Retrieve(region string, ec2Client ec2.Client) ([]string, error) {
	a.RetrieveCall.Receives.Region = region
	a.RetrieveCall.Receives.EC2Client = ec2Client
	return a.RetrieveCall.Returns.AZs, a.RetrieveCall.Returns.Error
}
