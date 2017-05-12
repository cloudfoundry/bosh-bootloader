package fakes

type VPCStatusChecker struct {
	ValidateSafeToDeleteCall struct {
		CallCount int
		Receives  struct {
			VPCID string
			EnvID string
		}
		Returns struct {
			Error error
		}
	}
}

func (v *VPCStatusChecker) ValidateSafeToDelete(vpcID, envID string) error {
	v.ValidateSafeToDeleteCall.CallCount++
	v.ValidateSafeToDeleteCall.Receives.VPCID = vpcID
	v.ValidateSafeToDeleteCall.Receives.EnvID = envID
	return v.ValidateSafeToDeleteCall.Returns.Error
}
