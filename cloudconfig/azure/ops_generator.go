package azure

import (
	// "fmt"
	"strings"

	yaml "gopkg.in/yaml.v2"

	// "github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type OpsGenerator struct {
	terraformManager terraformManager
}

type terraformManager interface {
	GetOutputs(storage.State) (map[string]interface{}, error)
}

type op struct {
	Type  string
	Path  string
	Value interface{}
}

var marshal func(interface{}) ([]byte, error) = yaml.Marshal

func NewOpsGenerator(terraformManager terraformManager) OpsGenerator {
	return OpsGenerator{
		terraformManager: terraformManager,
	}
}

func (o OpsGenerator) Generate(state storage.State) (string, error) {
	// ops, err := [], nil
	// if err != nil {
	// 	return "", err
	// }
	ops := []op{}

	cloudConfigOpsYAML, err := marshal(ops)
	if err != nil {
		return "", err
	}

	return strings.Join(
		[]string{
			// BaseOps,
			string(cloudConfigOpsYAML),
		},
		"\n",
	), nil
}
