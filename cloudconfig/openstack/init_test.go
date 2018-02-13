package openstack

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestOpenStack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cloudconfig/openstack")
}
