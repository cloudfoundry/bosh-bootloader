package gcp_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCloudConfigGCP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "cloudconfig/gcp")
}
