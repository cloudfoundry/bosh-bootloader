package fakes

type ELBDescriber struct {
	DescribeCall struct {
		CallCount int
		Stub      func(string) ([]string, error)
		Receives  struct {
			ElbName string
		}
		Returns struct {
			Instances []string
			Error     error
		}
	}
}

func (e *ELBDescriber) Describe(elbName string) ([]string, error) {
	e.DescribeCall.CallCount++
	e.DescribeCall.Receives.ElbName = elbName

	if e.DescribeCall.Stub != nil {
		return e.DescribeCall.Stub(elbName)
	}

	return e.DescribeCall.Returns.Instances, e.DescribeCall.Returns.Error
}
