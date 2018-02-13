package fakes

import (
	"io"
)

type BOSHCommand struct {
	GetBOSHPathCall struct {
		CallCount int
		Returns   struct {
			Path  string
			Error error
		}
	}

	RunStub        func(stdout io.Writer, args []string) error
	runArgsForCall []struct {
		stdout io.Writer
		args   []string
	}
	runReturns struct {
		result1 error
	}
	runReturnsOnCall map[int]struct {
		result1 error
	}
	invocations map[string][][]interface{}
}

func (fake *BOSHCommand) GetBOSHPath() (string, error) {
	fake.GetBOSHPathCall.CallCount++

	return fake.GetBOSHPathCall.Returns.Path, fake.GetBOSHPathCall.Returns.Error
}

func (fake *BOSHCommand) Run(stdout io.Writer, args []string) error {
	var argsCopy []string
	if args != nil {
		argsCopy = make([]string, len(args))
		copy(argsCopy, args)
	}
	ret, specificReturn := fake.runReturnsOnCall[len(fake.runArgsForCall)]
	fake.runArgsForCall = append(fake.runArgsForCall, struct {
		stdout io.Writer
		args   []string
	}{stdout, argsCopy})
	fake.recordInvocation("Run", []interface{}{stdout, argsCopy})

	if fake.RunStub != nil {
		return fake.RunStub(stdout, args)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.runReturns.result1
}

func (fake *BOSHCommand) RunCallCount() int {
	return len(fake.runArgsForCall)
}

func (fake *BOSHCommand) RunArgsForCall(i int) (io.Writer, []string) {
	return fake.runArgsForCall[i].stdout, fake.runArgsForCall[i].args
}

func (fake *BOSHCommand) RunReturns(result1 error) {
	fake.RunStub = nil
	fake.runReturns = struct {
		result1 error
	}{result1}
}

func (fake *BOSHCommand) RunReturnsOnCall(i int, result1 error) {
	fake.RunStub = nil
	if fake.runReturnsOnCall == nil {
		fake.runReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.runReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *BOSHCommand) Invocations() map[string][][]interface{} {
	return fake.invocations
}

func (fake *BOSHCommand) recordInvocation(key string, args []interface{}) {
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}
