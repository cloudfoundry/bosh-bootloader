package awsbackend

import (
	"sync"

	"github.com/rosenhouse/awsfaker"
)

type Stack struct {
	Name       string
	Template   string
	WasUpdated bool
}

type Stacks struct {
	mutex sync.Mutex
	store map[string]Stack

	createStack struct {
		returns struct {
			err *awsfaker.ErrorResponse
		}
	}

	deleteStack struct {
		returns struct {
			err *awsfaker.ErrorResponse
		}
	}
}

func NewStacks() *Stacks {
	return &Stacks{
		store: make(map[string]Stack),
	}
}

func (s *Stacks) Set(stack Stack) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store[stack.Name] = stack
}

func (s *Stacks) Get(name string) (Stack, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	stack, ok := s.store[name]
	return stack, ok
}

func (s *Stacks) Delete(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.store, name)
}

func (s *Stacks) SetCreateStackReturnError(err *awsfaker.ErrorResponse) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.createStack.returns.err = err
}

func (s *Stacks) CreateStackReturnError() *awsfaker.ErrorResponse {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.createStack.returns.err
}

func (s *Stacks) SetDeleteStackReturnError(err *awsfaker.ErrorResponse) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.deleteStack.returns.err = err
}

func (s *Stacks) DeleteStackReturnError() *awsfaker.ErrorResponse {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.deleteStack.returns.err
}
