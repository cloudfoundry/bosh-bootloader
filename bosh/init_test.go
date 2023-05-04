package bosh_test

import (
	"os"
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestBosh(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "bosh")
}

var (
	originalPath string
)

var _ = BeforeSuite(func() {
	originalPath = os.Getenv("PATH")
})
