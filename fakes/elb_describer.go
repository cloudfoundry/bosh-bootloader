package fakes

import "github.com/pivotal-cf-experimental/bosh-bootloader/aws/elb"

type ELBDescriber struct {
	DescribeCall struct {
		CallCount int
		Stub      func(string) ([]string, error)
		Receives  struct {
			ElbName string
			Client  elb.Client
		}
		Returns struct {
			Instances []string
			Error     error
		}
	}
}

func (e *ELBDescriber) Describe(elbName string, client elb.Client) ([]string, error) {
	e.DescribeCall.CallCount++
	e.DescribeCall.Receives.ElbName = elbName
	e.DescribeCall.Receives.Client = client

	if e.DescribeCall.Stub != nil {
		return e.DescribeCall.Stub(elbName)
	}

	return e.DescribeCall.Returns.Instances, e.DescribeCall.Returns.Error
}
