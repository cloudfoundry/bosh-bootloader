package prettyprint_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/prettyprint"
)

var _ = Describe("SliceAsValue", func() {
	It("returns a string representation of the slice", func() {
		sliceAsValue := reflect.ValueOf([]string{"a", "b"})
		Expect(prettyprint.SliceAsValue(sliceAsValue)).To(Equal("[<string> a, <string> b]"))
	})
})
