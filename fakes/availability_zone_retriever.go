package fakes

type AvailabilityZoneRetriever struct {
	RetrieveAvailabilityZonesCall struct {
		Receives struct {
			Region string
		}
		Returns struct {
			AZs   []string
			Error error
		}
		CallCount int
	}
}

func (a *AvailabilityZoneRetriever) RetrieveAvailabilityZones(region string) ([]string, error) {
	a.RetrieveAvailabilityZonesCall.Receives.Region = region
	a.RetrieveAvailabilityZonesCall.CallCount++
	return a.RetrieveAvailabilityZonesCall.Returns.AZs, a.RetrieveAvailabilityZonesCall.Returns.Error
}
