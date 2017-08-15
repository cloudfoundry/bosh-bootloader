package azure_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestAzure(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "terraform/azure")
}
