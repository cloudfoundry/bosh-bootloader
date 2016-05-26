package awsbackend

import "sync"

type Certificate struct {
	Name            string
	CertificateBody string
	PrivateKey      string
	Chain           string
}

type Certificates struct {
	mutex sync.Mutex
	store map[string]Certificate
}

func NewCertificates() *Certificates {
	return &Certificates{
		store: make(map[string]Certificate),
	}
}

func (c *Certificates) Set(certificate Certificate) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.store[certificate.Name] = certificate
}

func (c *Certificates) Get(name string) (Certificate, bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	certificate, ok := c.store[name]
	return certificate, ok
}

func (c *Certificates) Delete(name string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.store, name)
}

func (c *Certificates) All() []Certificate {
	var certificates []Certificate

	c.mutex.Lock()
	defer c.mutex.Unlock()
	for _, certificate := range c.store {
		certificates = append(certificates, certificate)
	}

	return certificates
}
