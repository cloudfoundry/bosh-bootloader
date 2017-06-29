package multierror_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestMultierror(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Multierror Suite")
}
