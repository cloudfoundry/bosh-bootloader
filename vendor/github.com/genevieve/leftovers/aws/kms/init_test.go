package kms_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestKms(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "aws/kms")
}
