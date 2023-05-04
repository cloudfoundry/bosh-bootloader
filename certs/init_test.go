package certs

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestIAM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "certs")
}
