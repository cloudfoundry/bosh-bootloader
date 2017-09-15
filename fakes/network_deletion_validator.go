package fakes

type NetworkDeletionValidator struct {
	ValidateSafeToDeleteCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
		Receives struct {
			NetworkName string
			EnvID       string
		}
	}
}

func (n *NetworkDeletionValidator) ValidateSafeToDelete(networkName string, envID string) error {
	n.ValidateSafeToDeleteCall.CallCount++
	n.ValidateSafeToDeleteCall.Receives.NetworkName = networkName
	n.ValidateSafeToDeleteCall.Receives.EnvID = envID

	return n.ValidateSafeToDeleteCall.Returns.Error
}
