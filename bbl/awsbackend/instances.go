package awsbackend

import "sync"

type Instance struct {
	Name  string
	VPCID string
}

type Instances struct {
	mutex sync.Mutex
	store []Instance
}

func NewInstances() *Instances {
	return &Instances{}
}

func (i *Instances) Set(instances []Instance) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	i.store = instances
}

func (i *Instances) Get() []Instance {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	return i.store
}
