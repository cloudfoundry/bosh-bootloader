package fakes

type AvailabilityZoneRetriever struct {
	RetrieveCall struct {
		Receives struct {
			Region string
		}
		Returns struct {
			AZs   []string
			Error error
		}
	}
}

func (a *AvailabilityZoneRetriever) Retrieve(region string) ([]string, error) {
	a.RetrieveCall.Receives.Region = region
	return a.RetrieveCall.Returns.AZs, a.RetrieveCall.Returns.Error
}
