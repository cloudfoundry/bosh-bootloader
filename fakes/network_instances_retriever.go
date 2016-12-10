package fakes

type NetworkInstancesRetriever struct {
	ListCall struct {
		Returns struct {
			Instances []string
			Error     error
		}

		Receives struct {
			ProjectID   string
			Zone        string
			NetworkName string
		}
	}
}

func (n *NetworkInstancesRetriever) List(projectID, zone, networkName string) ([]string, error) {
	n.ListCall.Receives.ProjectID = projectID
	n.ListCall.Receives.Zone = zone
	n.ListCall.Receives.NetworkName = networkName

	return n.ListCall.Returns.Instances, n.ListCall.Returns.Error
}
