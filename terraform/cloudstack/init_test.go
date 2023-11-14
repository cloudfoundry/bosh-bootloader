package cloudstack_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestCloudStack(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "terraform/vsphere")
}
