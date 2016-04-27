package awsbackend

import "sync"

type LoadBalancer struct {
	Name      string
	Instances []string
}

type LoadBalancers struct {
	mutex sync.Mutex
	store map[string]LoadBalancer
}

func NewLoadBalancers() *LoadBalancers {
	return &LoadBalancers{
		store: make(map[string]LoadBalancer),
	}
}

func (s *LoadBalancers) Set(loadBalancer LoadBalancer) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.store[loadBalancer.Name] = loadBalancer
}

func (s *LoadBalancers) Get(name string) (LoadBalancer, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	loadBalancer, ok := s.store[name]
	return loadBalancer, ok
}

func (s *LoadBalancers) Delete(name string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.store, name)
}
