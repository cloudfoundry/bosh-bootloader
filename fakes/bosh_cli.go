package fakes

import (
	"io"
	"sync"
)

type BOSHCLI struct {
	GetBOSHPathCall struct {
		CallCount int
		Returns   struct {
			Path string
		}
	}
	RunCall struct {
		CallCount int
		Receieves struct {
			Stdout           io.Writer
			WorkingDirectory string
			Args             []string
		}
		Returns struct {
			Error error
		}
	}
	UpdateRuntimeConfigCall struct {
		CallCount int
		Receives  struct {
			Filepath, Name string
		}
		Returns struct {
			Error error
		}
	}
	RunStub        func(stdout io.Writer, workingDirectory string, args []string) error
	runMutex       sync.RWMutex
	runArgsForCall []struct {
		stdout           io.Writer
		workingDirectory string
		args             []string
	}
	runReturns struct {
		result1 error
	}
	runReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *BOSHCLI) GetBOSHPath() string {
	fake.GetBOSHPathCall.CallCount++

	return fake.GetBOSHPathCall.Returns.Path
}

func (fake *BOSHCLI) Run(stdout io.Writer, workingDirectory string, args []string) error {
	var argsCopy []string
	if args != nil {
		argsCopy = make([]string, len(args))
		copy(argsCopy, args)
	}
	fake.runMutex.Lock()
	ret, specificReturn := fake.runReturnsOnCall[len(fake.runArgsForCall)]
	fake.runArgsForCall = append(fake.runArgsForCall, struct {
		stdout           io.Writer
		workingDirectory string
		args             []string
	}{stdout, workingDirectory, argsCopy})
	fake.recordInvocation("Run", []interface{}{stdout, workingDirectory, argsCopy})
	fake.runMutex.Unlock()

	if fake.RunStub != nil {
		return fake.RunStub(stdout, workingDirectory, args)
	}
	if specificReturn {
		return ret.result1
	}
	return fake.runReturns.result1
}

func (fake *BOSHCLI) RunCallCount() int {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return len(fake.runArgsForCall)
}

func (fake *BOSHCLI) RunArgsForCall(i int) (io.Writer, string, []string) {
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return fake.runArgsForCall[i].stdout, fake.runArgsForCall[i].workingDirectory, fake.runArgsForCall[i].args
}

func (fake *BOSHCLI) RunReturns(result1 error) {
	fake.RunStub = nil
	fake.runReturns = struct {
		result1 error
	}{result1}
}

func (fake *BOSHCLI) RunReturnsOnCall(i int, result1 error) {
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

func (fake *BOSHCLI) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.runMutex.RLock()
	defer fake.runMutex.RUnlock()
	return fake.invocations
}

func (fake *BOSHCLI) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

func (fake *BOSHCLI) UpdateRuntimeConfig(filepath, name string) error {
	fake.UpdateRuntimeConfigCall.CallCount++

	fake.UpdateRuntimeConfigCall.Receives.Filepath = filepath
	fake.UpdateRuntimeConfigCall.Receives.Name = name

	return fake.UpdateRuntimeConfigCall.Returns.Error
}
