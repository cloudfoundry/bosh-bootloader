package bosh_test

import (
	"testing"

	"github.com/pivotal-cf-experimental/bosh-test/bosh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestBosh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bosh")
}

var _ = AfterEach(func() {
	bosh.ResetBodyReader()
})
