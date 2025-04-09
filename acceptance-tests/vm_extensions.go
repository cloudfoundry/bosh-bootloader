package acceptance

import (
	"gopkg.in/yaml.v2"

	. "github.com/onsi/gomega" //nolint:staticcheck
)

func VmExtensionNames(cloudConfigOutput string) []string {
	var cloudConfig struct {
		VMExtensions []struct {
			Name            string                 `yaml:"name"`
			CloudProperties map[string]interface{} `yaml:"cloud_properties"`
		} `yaml:"vm_extensions"`
	}
	err := yaml.Unmarshal([]byte(cloudConfigOutput), &cloudConfig)
	Expect(err).NotTo(HaveOccurred())

	var names []string
	for _, extension := range cloudConfig.VMExtensions {
		names = append(names, extension.Name)
	}
	return names
}
