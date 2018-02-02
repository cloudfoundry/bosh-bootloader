package ec2_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestEC2(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "aws/ec2")
}
