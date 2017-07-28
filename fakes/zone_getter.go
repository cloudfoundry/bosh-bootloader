package fakes

type Zones struct {
	GetCall struct {
		CallCount int
		Receives  struct {
			Region string
		}
		Returns struct {
			Zones []string
			Error error
		}
	}
}

func (z *Zones) Get(region string) ([]string, error) {
	z.GetCall.CallCount++
	z.GetCall.Receives.Region = region
	return z.GetCall.Returns.Zones, z.GetCall.Returns.Error
}
