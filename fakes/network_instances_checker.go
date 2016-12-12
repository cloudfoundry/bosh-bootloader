package fakes

type NetworkInstancesChecker struct {
	ValidateSafeToDeleteCall struct {
		Returns struct {
			Error error
		}

		Receives struct {
			ProjectID   string
			Zone        string
			NetworkName string
		}
	}
}

func (n *NetworkInstancesChecker) ValidateSafeToDelete(projectID, zone, networkName string) error {
	n.ValidateSafeToDeleteCall.Receives.ProjectID = projectID
	n.ValidateSafeToDeleteCall.Receives.Zone = zone
	n.ValidateSafeToDeleteCall.Receives.NetworkName = networkName

	return n.ValidateSafeToDeleteCall.Returns.Error
}
