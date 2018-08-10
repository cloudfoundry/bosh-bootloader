package renderers_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestRenderers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "renderers")
}
