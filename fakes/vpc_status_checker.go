package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"

type VPCStatusChecker struct {
	ValidateSafeToDeleteCall struct {
		Receives struct {
			Client ec2.Client
			VPCID  string
		}
		Returns struct {
			Error error
		}
	}
}

func (v *VPCStatusChecker) ValidateSafeToDelete(client ec2.Client, vpcID string) error {
	v.ValidateSafeToDeleteCall.Receives.Client = client
	v.ValidateSafeToDeleteCall.Receives.VPCID = vpcID
	return v.ValidateSafeToDeleteCall.Returns.Error
}
