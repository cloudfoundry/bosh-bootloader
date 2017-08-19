package actors

import (
	"net/http"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
)

type BOSH struct{}

func NewBOSH() BOSH {
	return BOSH{}
}

func (BOSH) DirectorExists(address, username, password string) bool {
	client := bosh.NewClient(http.DefaultClient, false, address, username, password, "")

	_, err := client.Info()
	return err == nil
}
