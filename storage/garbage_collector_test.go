package storage_test

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("garbage collector", func() {
	var (
		gc     storage.GarbageCollector
		fileIO *fakes.FileIO
	)

	BeforeEach(func() {
		fileIO = &fakes.FileIO{}
		fileIO.StatCall.Returns.FileInfo = &fakes.FileInfo{}
		gc = storage.NewGarbageCollector(fileIO)
	})

	Describe("remove", func() {
		It("removes the bbl-state.json file", func() {
			err := gc.Remove("some-dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fileIO.RemoveCall.Receives[0].Name).To(Equal(filepath.Join("some-dir", "bbl-state.json")))
			Expect(fileIO.RemoveCall.Receives[0].Name).To(Equal(filepath.Join("some-dir", "bbl-state.json")))
		})

		It("doesn't remove user managed files", func() {
			err := gc.Remove("some-dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fileIO.RemoveCall.Receives).NotTo(ContainElement(
				fakes.RemoveReceive{Name: filepath.Join("some-dir", "create-jumpbox-override.sh")},
			))
			Expect(fileIO.RemoveAllCall.Receives).NotTo(ContainElement(
				fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "create-jumpbox-override.sh")},
			))
		})

		It("removes bosh *-env scripts", func() {
			createDirector := filepath.Join("some-dir", "create-director.sh")
			createJumpbox := filepath.Join("some-dir", "create-jumpbox.sh")
			deleteDirector := filepath.Join("some-dir", "delete-director.sh")
			deleteJumpbox := filepath.Join("some-dir", "delete-jumpbox.sh")

			err := gc.Remove("some-dir")
			Expect(err).NotTo(HaveOccurred())

			Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(fakes.RemoveAllReceive{Path: createDirector}))
			Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(fakes.RemoveAllReceive{Path: deleteDirector}))
			Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(fakes.RemoveAllReceive{Path: deleteJumpbox}))
			Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(fakes.RemoveAllReceive{Path: createJumpbox}))
		})

		DescribeTable("removing bbl-created directories",
			func(directory string, expectToBeDeleted bool) {
				err := gc.Remove("some-dir")
				Expect(err).NotTo(HaveOccurred())

				if expectToBeDeleted {
					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(fakes.RemoveAllReceive{
						Path: filepath.Join("some-dir", directory),
					}))
				} else {
					Expect(fileIO.RemoveAllCall.Receives).NotTo(ContainElement(fakes.RemoveAllReceive{
						Path: filepath.Join("some-dir", directory),
					}))
				}
			},
			Entry(".terraform", ".terraform", true),
			Entry("bosh-deployment", "bosh-deployment", true),
			Entry("jumpbox-deployment", "jumpbox-deployment", true),
			Entry("bbl-ops-files", "bbl-ops-files", true),
			Entry("non-bbl directory", "foo", false),
		)

		Describe("cloud-config", func() {
			var (
				cloudConfigBase string
				cloudConfigOps  string
			)
			BeforeEach(func() {
				cloudConfigBase = filepath.Join("some-dir", "cloud-config", "cloud-config.yml")
				cloudConfigOps = filepath.Join("some-dir", "cloud-config", "ops.yml")
				fileIO.StatCall.Returns.FileInfo = &fakes.DirFileInfo{}
			})

			It("removes the ops file and base file; removes the directory if there are no user-provided files", func() {
				err := gc.Remove("some-dir")
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(fakes.RemoveAllReceive{Path: cloudConfigBase}))
				Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(fakes.RemoveAllReceive{Path: cloudConfigOps}))
				// don't remove populated, relevant dirs
				Expect(fileIO.RemoveCall.Receives).To(ContainElement(fakes.RemoveReceive{
					Name: filepath.Join("some-dir", "cloud-config"),
				}))
			})
		})

		Describe("runtime-config", func() {
			var runtimeConfig string

			BeforeEach(func() {
				runtimeConfig = filepath.Join("some-dir", "runtime-config", "runtime-config.yml")
				fileIO.StatCall.Returns.FileInfo = &fakes.DirFileInfo{}
			})

			It("removes the runtime-config file; removes the directory if there are no user-provided files", func() {
				err := gc.Remove("some-dir")
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(fakes.RemoveAllReceive{Path: runtimeConfig}))
				// don't remove populated, relevant dirs
				Expect(fileIO.RemoveCall.Receives).To(ContainElement(fakes.RemoveReceive{
					Name: filepath.Join("some-dir", "runtime-config"),
				}))
			})
		})

		Describe("vars", func() {
			Context("when the vars directory contains only bbl files", func() {
				BeforeEach(func() {
					fileIO.ReadDirCall.Returns.FileInfos = []os.FileInfo{
						fakes.FileInfo{FileName: "bbl.tfvars"},
						fakes.FileInfo{FileName: "bosh-state.json"},
						fakes.FileInfo{FileName: "cloud-config-vars.yml"},
						fakes.FileInfo{FileName: "director-vars-file.yml"},
						fakes.FileInfo{FileName: "director-vars-store.yml"},
						fakes.FileInfo{FileName: "jumpbox-state.json"},
						fakes.FileInfo{FileName: "jumpbox-vars-file.yml"},
						fakes.FileInfo{FileName: "jumpbox-vars-store.yml"},
						fakes.FileInfo{FileName: "terraform.tfstate"},
						fakes.FileInfo{FileName: "terraform.tfstate.backup"},
					}
					fileIO.StatCall.Returns.FileInfo = &fakes.DirFileInfo{}
				})

				It("removes the directory and its contents", func() {
					err := gc.Remove("some-dir")
					Expect(err).NotTo(HaveOccurred())

					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(
						fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "vars", "bbl.tfvars")},
					))
					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(
						fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "vars", "bosh-state.json")},
					))
					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(
						fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "vars", "cloud-config-vars.yml")},
					))
					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(
						fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "vars", "director-vars-file.yml")},
					))
					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(
						fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "vars", "director-vars-store.yml")},
					))
					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(
						fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "vars", "jumpbox-state.json")},
					))
					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(
						fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "vars", "jumpbox-vars-file.yml")},
					))
					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(
						fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "vars", "jumpbox-vars-store.yml")},
					))
					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(
						fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "vars", "terraform.tfstate")},
					))
					Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(
						fakes.RemoveAllReceive{Path: filepath.Join("some-dir", "vars", "terraform.tfstate.backup")},
					))
					Expect(fileIO.RemoveCall.Receives).To(ContainElement(fakes.RemoveReceive{
						Name: filepath.Join("some-dir", "vars"),
					}))
				})
			})

			Context("when the vars directory contains user managed files", func() {
				BeforeEach(func() {
					fileIO.ReadDirCall.Returns.FileInfos = []os.FileInfo{
						fakes.FileInfo{FileName: "user-managed-file"},
						fakes.FileInfo{FileName: "terraform.tfstate.backup"},
					}
				})

				It("spares user managed files", func() {
					err := gc.Remove("some-dir")
					Expect(err).NotTo(HaveOccurred())

					Expect(fileIO.RemoveCall.Receives).NotTo(ContainElement(fakes.RemoveReceive{
						Name: filepath.Join("some-dir", "vars", "user-managed-file"),
					}))
				})
			})
		})

		Describe("terraform", func() {
			BeforeEach(func() {
				fileIO.StatCall.Returns.FileInfo = &fakes.DirFileInfo{}
			})

			It("removes the bbl template and directory", func() {
				bblTerraformTemplate := filepath.Join("some-dir", "terraform", "bbl-template.tf")

				err := gc.Remove("some-dir")
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.RemoveAllCall.Receives).To(ContainElement(fakes.RemoveAllReceive{Path: bblTerraformTemplate}))
				Expect(fileIO.RemoveCall.Receives).To(ContainElement(fakes.RemoveReceive{
					Name: filepath.Join("some-dir", "terraform"),
				}))
			})
		})

		Context("when the bbl-state.json file does not exist", func() {
			It("does nothing", func() {
				err := gc.Remove("some-dir")
				Expect(err).NotTo(HaveOccurred())

				Expect(len(fileIO.WriteFileCall.Receives)).To(Equal(0))
			})
		})

		Context("failure cases", func() {
			Context("when the bbl-state.json file cannot be removed", func() {
				BeforeEach(func() {
					fileIO.RemoveCall.Returns = []fakes.RemoveReturn{{Error: errors.New("permission denied")}}
				})

				It("returns an error", func() {
					err := gc.Remove("some-dir")
					Expect(err).To(MatchError(ContainSubstring("permission denied")))
				})
			})
		})
	})
})
