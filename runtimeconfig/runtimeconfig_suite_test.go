package runtimeconfig_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRuntimeconfig(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Runtimeconfig Suite")
}
