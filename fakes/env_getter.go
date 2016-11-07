package fakes

type EnvGetter struct {
	Values  map[string]string
	GetCall struct {
		CallCount int
		Receives  struct {
			Name string
		}
	}
}

func (e *EnvGetter) Get(name string) string {
	e.GetCall.CallCount++
	return e.Values[name]
}
