package gcp_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestGCP(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "gcp")
}
