package commands_test

import (
	"errors"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Up", func() {
	var (
		command commands.Up

		fakeUp          *fakes.UpCmd
		fakeBOSHManager *fakes.BOSHManager
	)

	BeforeEach(func() {
		fakeUp = &fakes.UpCmd{}
		fakeBOSHManager = &fakes.BOSHManager{}
		fakeBOSHManager.VersionCall.Returns.Version = "2.0.24"

		command = commands.NewUp(fakeUp, fakeBOSHManager)
	})

	Describe("CheckFastFails", func() {
		Context("when the version of BOSH is a dev build", func() {
			It("does not fail", func() {
				fakeBOSHManager.VersionCall.Returns.Error = bosh.NewBOSHVersionError(errors.New("BOSH version could not be parsed"))

				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the version of BOSH is lower than 2.0.24", func() {
			It("returns a helpful error message when bbling up with a director", func() {
				fakeBOSHManager.VersionCall.Returns.Version = "1.9.1"
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err).To(MatchError("BOSH version must be at least v2.0.24"))
			})

			Context("when the no-director flag is specified", func() {
				It("does not return an error", func() {
					fakeBOSHManager.VersionCall.Returns.Version = "1.9.1"
					err := command.CheckFastFails([]string{
						"--no-director",
					}, storage.State{Version: 999})

					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		Context("when the version of BOSH cannot be retrieved", func() {
			It("returns an error", func() {
				fakeBOSHManager.VersionCall.Returns.Error = errors.New("BOOM")
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err.Error()).To(ContainSubstring("BOOM"))
			})
		})

		Context("when the version of BOSH is invalid", func() {
			It("returns an error", func() {
				fakeBOSHManager.VersionCall.Returns.Version = "X.5.2"
				err := command.CheckFastFails([]string{}, storage.State{Version: 999})

				Expect(err.Error()).To(ContainSubstring("invalid syntax"))
			})
		})

		Context("when bbl-state contains an env-id", func() {
			Context("when the passed in name matches the env-id", func() {
				It("returns no error", func() {
					err := command.CheckFastFails([]string{
						"--name", "some-name",
					}, storage.State{EnvID: "some-name"})
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("when the passed in name does not match the env-id", func() {
				It("returns an error", func() {
					err := command.CheckFastFails([]string{
						"--name", "some-other-name",
					}, storage.State{EnvID: "some-name"})
					Expect(err).To(MatchError("The director name cannot be changed for an existing environment. Current name is some-name."))
				})
			})
		})
	})

	Describe("Execute", func() {
		It("it works", func() {
			err := command.Execute([]string{}, storage.State{})
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeUp.ExecuteCall.CallCount).To(Equal(1))
		})

		Context("when the --ops-file flag is specified", func() {
			It("populates the aws config with the correct ops-file path", func() {
				err := command.Execute([]string{
					"--ops-file", "some-ops-file-path",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeUp.ExecuteCall.Receives.UpConfig.OpsFile).To(Equal("some-ops-file-path"))
			})

			Context("when the --ops-file flag is not specified", func() {
				It("creates a default ops-file with the contents of state.BOSH.UserOpsFile", func() {
					err := command.Execute([]string{}, storage.State{
						BOSH: storage.BOSH{
							UserOpsFile: "some-ops-file-contents",
						},
					})
					Expect(err).NotTo(HaveOccurred())

					filePath := fakeUp.ExecuteCall.Receives.UpConfig.OpsFile
					fileContents, err := ioutil.ReadFile(filePath)
					Expect(err).NotTo(HaveOccurred())

					Expect(string(fileContents)).To(Equal("some-ops-file-contents"))
				})
			})
		})

		Context("when the user provides the name flag", func() {
			It("passes the name flag in the up config", func() {
				err := command.Execute([]string{
					"--name", "a-better-name",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeUp.ExecuteCall.Receives.UpConfig.Name).To(Equal("a-better-name"))
			})
		})

		Context("when the user provides the no-director flag", func() {
			It("passes no-director as true in the up config", func() {
				err := command.Execute([]string{
					"--no-director",
				}, storage.State{})
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeUp.ExecuteCall.Receives.UpConfig.NoDirector).To(Equal(true))
			})

			Context("when the --no-director flag was omitted on a subsequent bbl-up", func() {
				It("passes no-director as true in the up config", func() {
					err := command.Execute([]string{},
						storage.State{
							IAAS:       "gcp",
							NoDirector: true,
						})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeUp.ExecuteCall.Receives.UpConfig.NoDirector).To(Equal(true))
				})
			})
		})

		Context("failure cases", func() {
			It("returns an error when the up command fails", func() {
				fakeUp.ExecuteCall.Returns.Error = errors.New("failed execution")
				err := command.Execute([]string{}, storage.State{})
				Expect(err).To(MatchError("failed execution"))
			})

			It("returns an error when undefined flags are passed", func() {
				err := command.Execute([]string{"--foo", "bar"}, storage.State{})
				Expect(err).To(MatchError("flag provided but not defined: -foo"))
			})
		})
	})
})
