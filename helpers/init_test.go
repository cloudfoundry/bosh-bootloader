package helpers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHelpers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "helpers")
}

func HasUniqueValues(values []string) bool {
	valueMap := make(map[string]struct{})

	for _, value := range values {
		valueMap[value] = struct{}{}
	}

	return len(valueMap) == len(values)
}
