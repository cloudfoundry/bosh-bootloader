package iam

import (
	"sync"
)

// Copied from github.com/hashicorp/terraform/helper
//
// MutexKV is a simple key/value store for arbitrary mutexes. It can be used to
// serialize changes across arbitrary collaborators that share knowledge of the
// keys they must serialize on.
//
// The initial use case is to let aws_security_group_rule resources serialize
// their access to individual security groups based on SG ID.
type MutexKV struct {
	lock  sync.Mutex
	store map[string]*sync.Mutex
}

// Locks the mutex for the given key. Caller is responsible for calling Unlock
// for the same key
func (m *MutexKV) Lock(key string) {
	m.get(key).Lock()
}

// Unlock the mutex for the given key. Caller must have called Lock for the same key first
func (m *MutexKV) Unlock(key string) {
	m.get(key).Unlock()
}

// Returns a mutex for the given key, no guarantee of its lock status
func (m *MutexKV) get(key string) *sync.Mutex {
	m.lock.Lock()
	defer m.lock.Unlock()
	mutex, ok := m.store[key]
	if !ok {
		mutex = &sync.Mutex{}
		m.store[key] = mutex
	}
	return mutex
}

// Returns a properly initalized MutexKV
func NewMutexKV() *MutexKV {
	return &MutexKV{
		store: make(map[string]*sync.Mutex),
	}
}
