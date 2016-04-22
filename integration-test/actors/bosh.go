package actors

import "github.com/pivotal-cf-experimental/bosh-bootloader/bosh"

type BOSH struct{}

func NewBOSH() BOSH {
	return BOSH{}
}

func (BOSH) DirectorExists(address, username, password string) bool {
	client := bosh.NewClient(address, username, password)

	_, err := client.Info()
	return err == nil
}
