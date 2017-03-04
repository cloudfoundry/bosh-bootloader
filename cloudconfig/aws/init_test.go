package aws

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAWSCloudConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cloudconfig/aws")
}
