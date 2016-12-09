package fakes

type Zones struct {
	GetCall struct {
		CallCount int
		Receives  struct {
			Region string
		}
		Returns struct {
			Zones []string
		}
	}
}

func (z *Zones) Get(region string) []string {
	z.GetCall.CallCount++
	z.GetCall.Receives.Region = region
	return z.GetCall.Returns.Zones
}
