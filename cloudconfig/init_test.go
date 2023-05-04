package cloudconfig

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCloudConfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cloudconfig")
}
