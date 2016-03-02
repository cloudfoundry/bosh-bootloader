package awsbackend

import "sync"

type Stack struct {
	Name       string
	WasUpdated bool
}

type Stacks struct {
	mutex sync.Mutex
	store map[string]Stack
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
