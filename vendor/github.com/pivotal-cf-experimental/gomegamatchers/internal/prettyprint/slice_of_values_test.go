package prettyprint_test

import (
	"reflect"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/gomegamatchers/internal/prettyprint"
)

var _ = Describe("SliceOfValues", func() {
	It("returns a string representation of the slice", func() {
		sliceOfValues := []reflect.Value{reflect.ValueOf("a"), reflect.ValueOf("b")}
		Expect(prettyprint.SliceOfValues(sliceOfValues)).To(Equal("[<string> a, <string> b]"))
	})
})
