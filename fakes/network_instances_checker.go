package fakes

type NetworkInstancesChecker struct {
	ValidateSafeToDeleteCall struct {
		Returns struct {
			Error error
		}
		Receives struct {
			NetworkName string
		}
	}
}

func (n *NetworkInstancesChecker) ValidateSafeToDelete(networkName string) error {
	n.ValidateSafeToDeleteCall.Receives.NetworkName = networkName

	return n.ValidateSafeToDeleteCall.Returns.Error
}
