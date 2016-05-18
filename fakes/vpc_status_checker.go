package fakes

type VPCStatusChecker struct {
	ValidateSafeToDeleteCall struct {
		Receives struct {
			VPCID string
		}
		Returns struct {
			Error error
		}
	}
}

func (v *VPCStatusChecker) ValidateSafeToDelete(vpcID string) error {
	v.ValidateSafeToDeleteCall.Receives.VPCID = vpcID
	return v.ValidateSafeToDeleteCall.Returns.Error
}
