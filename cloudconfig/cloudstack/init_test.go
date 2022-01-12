package cloudstack

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCloudstack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cloudconfig/cloudstack")
}
