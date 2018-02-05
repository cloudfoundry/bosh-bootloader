package iam_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestIAM(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "aws/iam")
}
