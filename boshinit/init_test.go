package boshinit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestBOSHInit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "boshinit")
}
