package terraform_test

import (
	"errors"
	"os"
	"time"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	"github.com/cloudfoundry/bosh-bootloader/terraform/binary_dist"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("BinaryPath", func() {
	var (
		fileSystem *fakes.FileIO
	)

	BeforeEach(func() {
		fileSystem = &fakes.FileIO{}
		fileSystem.GetTempDirCall.Returns.Name = "/some/tmp/path/"
		fileSystem.ExistsCall.Returns.Bool = false
	})

	It("installs the binary from binary-dist and preserves modified timestamps", func() {
		res, err := terraform.BinaryPathInjected(fileSystem)
		Expect(err).NotTo(HaveOccurred())
		Expect(res).To(Equal("/some/tmp/path/bbl-terraform"))
		Expect(fileSystem.WriteFileCall.Receives[0].Filename).To(Equal("/some/tmp/path/bbl-terraform"))
		Expect(fileSystem.WriteFileCall.Receives[0].Mode).To(Equal(os.ModePerm))
		// this covers a critical corner case where the user installs
		// an old version AFTER we updated the tf binary that we distribute.
		Expect(fileSystem.ChtimesCall.Receives.ModTime).To(Equal(binary_dist.MustAssetInfo("terraform").ModTime()))
	})

	Context("when there's an error trying to find the cached bbl binary", func() {
		BeforeEach(func() {
			fileSystem.ExistsCall.Returns.Error = errors.New("bananananana")
		})

		It("errors", func() {
			_, err := terraform.BinaryPathInjected(fileSystem)
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
				_, err := terraform.BinaryPathInjected(fileSystem)
				Expect(err).To(MatchError(fileSystem.StatCall.Returns.Error))
			})
		})

		Context("that has the same modified timestamp as our binary dist", func() {
			BeforeEach(func() {
				info, err := binary_dist.AssetInfo("terraform")
				Expect(err).NotTo(HaveOccurred())
				modtime := info.ModTime()

				fileSystem.StatCall.Returns.FileInfo = fakes.FileInfo{
					Modtime: &modtime,
				}
			})

			It("doesn't rewrite the file", func() {
				res, err := terraform.BinaryPathInjected(fileSystem)
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
				res, err := terraform.BinaryPathInjected(fileSystem)
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
					_, err := terraform.BinaryPathInjected(fileSystem)
					Expect(err).To(MatchError(fileSystem.WriteFileCall.Returns[0].Error))
				})
			})

			Context("but we fail to set its timestamp", func() {
				BeforeEach(func() {
					fileSystem.ChtimesCall.Returns.Error = errors.New("banananann")
				})

				It("errors", func() {
					_, err := terraform.BinaryPathInjected(fileSystem)
					Expect(err).To(MatchError(fileSystem.ChtimesCall.Returns.Error))
				})
			})
		})
	})
})
