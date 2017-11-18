package vsphere

import "github.com/cloudfoundry/bosh-bootloader/storage"

type InputGenerator struct {
}

func NewInputGenerator() InputGenerator {
	return InputGenerator{}
}

func (i InputGenerator) Generate(state storage.State) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
