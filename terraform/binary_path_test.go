package terraform_test

import (
	"embed"
	"errors"
	"os"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//go:embed testassets/success
var content embed.FS

//go:embed testassets/only_mod_time
var contentOnlyModTime embed.FS

//go:embed testassets/only_terraform
var contentOnlyTerraform embed.FS

//go:embed testassets/incorrect_mod_time
var contentIncorrectModTime embed.FS

var _ = Describe("BinaryPath", func() {
	var (
		fileSystem *fakes.FileIO

		binary *terraform.Binary
	)

	BeforeEach(func() {
		path := "/some/tmp/path"
		fileSystem = &fakes.FileIO{}
		fileSystem.GetTempDirCall.Returns.Name = path
		fileSystem.ExistsCall.Returns.Bool = false

		binary = &terraform.Binary{
			Path:            "testassets/success",
			EmbedData:       content,
			FS:              fileSystem,
			TerraformBinary: "",
		}
	})

	It("installs the binary from binary-dist and preserves modified timestamps", func() {
		res, err := binary.BinaryPath()
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal("/some/tmp/path/bbl-terraform"))
		Expect(fileSystem.WriteFileCall.Receives[0].Filename).To(Equal("/some/tmp/path/bbl-terraform"))
		Expect(fileSystem.WriteFileCall.Receives[0].Mode).To(Equal(os.ModePerm))

		modTime, err := binary.RetrieveModTime()
		Expect(err).NotTo(HaveOccurred())
		// this covers a critical corner case where the user installs
		// an old version AFTER we updated the tf binary that we distribute.
		Expect(fileSystem.ChtimesCall.Receives.ModTime).To(Equal(modTime))
	})

	Context("when a custom terraform binary path is set", func() {
		BeforeEach(func() {
			binary.TerraformBinary = "/some/custom/path/bbl-terraform"
		})

		Context("and the file exists", func() {
			BeforeEach(func() {
				fileSystem.ExistsCall.Returns.Bool = true
			})

			It("doesn't rewrite the file", func() {
				res, err := binary.BinaryPath()
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal("/some/custom/path/bbl-terraform"))
				Expect(fileSystem.WriteFileCall.CallCount).To(Equal(0))
			})
		})

		Context("but the file does not exist", func() {
			It("uses the embedded binary ", func() {
				res, err := binary.BinaryPath()
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal("/some/tmp/path/bbl-terraform"))
			})
		})
	})

	Context("when there is no my-terraform-binary in box", func() {
		BeforeEach(func() {
			binary.EmbedData = contentOnlyModTime
			binary.Path = "testassets/only_mod_time"
		})

		It("returns an error", func() {
			_, err := binary.BinaryPath()
			Expect(err).To(MatchError(ContainSubstring("missing terraform")))
		})
	})

	Context("when there is no terraform-mod-time in box", func() {
		BeforeEach(func() {
			binary.EmbedData = contentOnlyTerraform
			binary.Path = "testassets/only_terraform"
		})

		It("returns an error", func() {
			_, err := binary.BinaryPath()
			Expect(err).To(MatchError("could not find terraform-mod-time file"))
		})
	})

	Context("when terraform mod time is formatted incorrectly", func() {
		BeforeEach(func() {
			binary.EmbedData = contentIncorrectModTime
			binary.Path = "testassets/incorrect_mod_time"
		})

		It("returns an error", func() {
			_, err := binary.BinaryPath()
			Expect(err).To(MatchError(ContainSubstring("incorrect format of time in terraform-mod-time")))
		})
	})

	Context("when there's an error trying to find the cached bbl binary", func() {
		BeforeEach(func() {
			fileSystem.ExistsCall.Returns.Error = errors.New("bananananana")
		})

		It("errors", func() {
			_, err := binary.BinaryPath()
			Expect(err).To(MatchError(fileSystem.ExistsCall.Returns.Error))
		})
	})

	// caching
	Context("when there is already a $TMP/bbl-terraform", func() {
		BeforeEach(func() {
			fileSystem.ExistsCall.Returns.Bool = true
		})

		Context("but we fail to stat it", func() {
			BeforeEach(func() {
				fileSystem.StatCall.Returns.Error = errors.New("banan")
			})

			It("errors", func() {
				_, err := binary.BinaryPath()
				Expect(err).To(MatchError(fileSystem.StatCall.Returns.Error))
			})
		})

		Context("that has the same modified timestamp as our binary dist", func() {
			BeforeEach(func() {
				modTime, err := binary.RetrieveModTime()
				Expect(err).NotTo(HaveOccurred())

				fileSystem.StatCall.Returns.FileInfo = fakes.FileInfo{
					Modtime: &modTime,
				}
			})

			It("doesn't rewrite the file", func() {
				res, err := binary.BinaryPath()
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal("/some/tmp/path/bbl-terraform"))
				Expect(fileSystem.WriteFileCall.CallCount).To(Equal(0))
			})
		})

		Context("that is old", func() {
			BeforeEach(func() {
				reallyOldTimestamp := time.Unix(0, 0)
				fileSystem.StatCall.Returns.FileInfo = fakes.FileInfo{
					Modtime: &reallyOldTimestamp,
				}
			})

			It("rewrites the file", func() {
				res, err := binary.BinaryPath()
				Expect(err).NotTo(HaveOccurred())
				Expect(res).To(Equal("/some/tmp/path/bbl-terraform"))
				Expect(fileSystem.WriteFileCall.CallCount).To(BeNumerically(">=", 1))
				Expect(fileSystem.WriteFileCall.Receives[0].Filename).To(Equal("/some/tmp/path/bbl-terraform"))
				Expect(fileSystem.WriteFileCall.Receives[0].Mode).To(Equal(os.ModePerm))
			})

			Context("but we fail to write it", func() {
				BeforeEach(func() {
					fileSystem.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("banananana")}}
				})

				It("errors", func() {
					_, err := binary.BinaryPath()
					Expect(err).To(MatchError(fileSystem.WriteFileCall.Returns[0].Error))
				})
			})

			Context("but we fail to set its timestamp", func() {
				BeforeEach(func() {
					fileSystem.ChtimesCall.Returns.Error = errors.New("banananann")
				})

				It("errors", func() {
					_, err := binary.BinaryPath()
					Expect(err).To(MatchError(fileSystem.ChtimesCall.Returns.Error))
				})
			})
		})
	})
})
