package fakes

type NetworkInstancesChecker struct {
	ValidateSafeToDeleteCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
		Receives struct {
			NetworkName string
		}
	}
}

func (n *NetworkInstancesChecker) ValidateSafeToDelete(networkName string) error {
	n.ValidateSafeToDeleteCall.CallCount++
	n.ValidateSafeToDeleteCall.Receives.NetworkName = networkName

	return n.ValidateSafeToDeleteCall.Returns.Error
}
