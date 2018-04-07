package cartographer_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestCartographer(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cartographer Suite")
}
