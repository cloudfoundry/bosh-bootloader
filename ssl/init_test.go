package ssl_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSSL(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ssl")
}
