package gomegamatchers_test

import (
	"errors"
	"os"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/gomegamatchers"
)

var _ = Describe("OsIsNotExistMatcher", func() {
	It("asserts if an error is an os.IsNotExist error", func() {
		Expect(os.ErrNotExist).To(gomegamatchers.BeAnOsIsNotExistError())
		Expect(errors.New("foo")).ToNot(gomegamatchers.BeAnOsIsNotExistError())
	})

	It("errors when not passed an error", func() {
		var foo error
		_, err := gomegamatchers.BeAnOsIsNotExistError().Match(foo)
		Expect(err).To(MatchError("BeAnOsIsNotExistError matcher expects an error, got <nil>"))
	})
})
