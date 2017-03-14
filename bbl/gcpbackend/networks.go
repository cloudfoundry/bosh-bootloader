package gcpbackend

import "sync"

type Network struct {
	networks []string
	mutex    sync.Mutex
}

func (n *Network) Add(network string) {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	n.networks = append(n.networks, network)
}

func (n *Network) Get() []string {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	return n.networks
}

func (n *Network) Exists(network string) bool {
	n.mutex.Lock()
	defer n.mutex.Unlock()
	for _, n := range n.networks {
		if n == network {
			return true
		}
	}
	return false
}
