package cloudconfig_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	var (
		logger           *fakes.Logger
		configUpdater    *fakes.BOSHConfigUpdater
		dirProvider      *fakes.DirProvider
		opsGenerator     *fakes.CloudConfigOpsGenerator
		terraformManager *fakes.TerraformManager
		fileIO           *fakes.FileIO
		manager          cloudconfig.Manager

		cloudConfigDir string
		varsDir        string
		incomingState  storage.State

		baseCloudConfig []byte
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		configUpdater = &fakes.BOSHConfigUpdater{}
		dirProvider = &fakes.DirProvider{}
		opsGenerator = &fakes.CloudConfigOpsGenerator{}
		terraformManager = &fakes.TerraformManager{}
		fileIO = &fakes.FileIO{}

		cloudConfigDir = "some-cloud-config-dir"
		dirProvider.GetCloudConfigDirCall.Returns.Directory = cloudConfigDir

		varsDir = "some-vars-dir"
		dirProvider.GetVarsDirCall.Returns.Directory = varsDir

		incomingState = storage.State{
			IAAS: "gcp",
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
		}

		opsGenerator.GenerateCall.Returns.OpsYAML = "some-ops"
		opsGenerator.GenerateVarsCall.Returns.VarsYAML = "some-vars"

		var err error
		baseCloudConfig, err = ioutil.ReadFile("fixtures/base-cloud-config.yml")
		Expect(err).NotTo(HaveOccurred())

		manager = cloudconfig.NewManager(logger, configUpdater, dirProvider, opsGenerator, terraformManager, fileIO)
	})

	Describe("Initialize", func() {
		It("returns a cloud config yaml with variable placeholders", func() {
			err := manager.Initialize(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join("some-cloud-config-dir", "cloud-config.yml")))
			Expect(fileIO.WriteFileCall.Receives[0].Contents).To(MatchYAML(baseCloudConfig))

			Expect(opsGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			Expect(fileIO.WriteFileCall.Receives[1].Filename).To(Equal(filepath.Join("some-cloud-config-dir", "ops.yml")))
			Expect(fileIO.WriteFileCall.Receives[1].Contents).To(Equal([]byte("some-ops")))
		})

		Context("failure cases", func() {
			Context("when getting the cloud config dir fails", func() {
				BeforeEach(func() {
					dirProvider.GetCloudConfigDirCall.Returns.Error = errors.New("carrot")
				})

				It("returns an error", func() {
					err := manager.Initialize(storage.State{})
					Expect(err).To(MatchError("carrot"))
				})
			})

			Context("when write file fails to write cloud-config.yml", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{
						Error: errors.New("apple"),
					}}
				})

				It("returns an error", func() {
					err := manager.Initialize(storage.State{})
					Expect(err).To(MatchError("apple"))
				})
			})

			Context("when ops generator fails to generate", func() {
				BeforeEach(func() {
					opsGenerator.GenerateCall.Returns.Error = errors.New("failed to generate")
				})

				It("returns an error", func() {
					err := manager.Initialize(storage.State{})
					Expect(err).To(MatchError("failed to generate"))
				})
			})

			Context("when write file fails to write the ops files", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{
						{
							Error: nil,
						},
						{
							Error: errors.New("banana"),
						},
					}
				})

				It("returns an error", func() {
					err := manager.Initialize(storage.State{})
					Expect(err).To(MatchError("banana"))
				})
			})
		})
	})

	Describe("IsPresentCloudConfig", func() {
		Context("when cloud config files exist in the state dir", func() {
			BeforeEach(func() {
				fileIO.StatCall.Returns.Error = nil
			})

			It("returns true", func() {
				Expect(manager.IsPresentCloudConfig()).To(BeTrue())
			})
		})

		Context("when base cloud config does not exist in the state dir", func() {
			BeforeEach(func() {
				fileIO.StatCall.Fake = func(name string) (os.FileInfo, error) {
					if strings.Contains(name, "cloud-config.yml") {
						return fakes.FileInfo{}, errors.New("nope")
					}
					return fakes.FileInfo{}, nil
				}
			})
			It("returns false", func() {
				Expect(manager.IsPresentCloudConfig()).To(BeFalse())
			})
		})

		Context("when bbl-defined ops file does not exist in the state dir", func() {
			BeforeEach(func() {
				fileIO.StatCall.Fake = func(name string) (os.FileInfo, error) {
					if strings.Contains(name, "ops.yml") {
						return fakes.FileInfo{}, errors.New("nope")
					}
					return fakes.FileInfo{}, nil
				}
			})
			It("returns false", func() {
				Expect(manager.IsPresentCloudConfig()).To(BeFalse())
			})
		})

		Context("failure cases", func() {
			Context("when getting the cloud config dir fails", func() {
				BeforeEach(func() {
					dirProvider.GetCloudConfigDirCall.Returns.Error = errors.New("carrot")
				})

				It("returns false", func() {
					Expect(manager.IsPresentCloudConfig()).To(BeFalse())
				})
			})
		})
	})

	Describe("IsPresentCloudConfigVars", func() {
		Context("when cloud config vars file exists in the vars dir", func() {
			BeforeEach(func() {
				fileIO.StatCall.Returns.Error = nil
			})

			It("returns true", func() {
				Expect(manager.IsPresentCloudConfigVars()).To(BeTrue())
			})
		})

		Context("when cloud config vars file does not exist in the vars dir", func() {
			BeforeEach(func() {
				fileIO.StatCall.Fake = func(name string) (os.FileInfo, error) {
					if strings.Contains(name, "cloud-config-vars.yml") {
						return fakes.FileInfo{}, errors.New("nope")
					}
					return fakes.FileInfo{}, nil
				}
			})

			It("returns false", func() {
				Expect(manager.IsPresentCloudConfigVars()).To(BeFalse())
			})
		})

		Context("failure cases", func() {
			Context("when getting the vars dir fails", func() {
				BeforeEach(func() {
					dirProvider.GetVarsDirCall.Returns.Error = errors.New("carrot")
				})

				It("returns false", func() {
					Expect(manager.IsPresentCloudConfigVars()).To(BeFalse())
				})
			})
		})
	})

	Describe("Update", func() {
		BeforeEach(func() {
			fileIO.ReadDirCall.Returns.FileInfos = []os.FileInfo{
				fakes.FileInfo{
					FileName: "shenanigans-ops.yml",
				},
				fakes.FileInfo{
					FileName: "ops.yml",
				},
				fakes.FileInfo{
					FileName: "cloud-config.yml",
				},
			}
		})

		It("logs steps taken", func() {
			err := manager.Update(incomingState)
			Expect(err).NotTo(HaveOccurred())
			Expect(logger.StepCall.Messages).To(Equal([]string{
				"generating cloud config",
				"applying cloud config",
			}))
		})

		It("updates the bosh director with a cloud config provided a valid bbl state", func() {
			boshCLI := bosh.AuthenticatedCLI{
				BOSHExecutablePath: "some-bosh-cli-path",
			}
			configUpdater.InitializeAuthenticatedCLICall.Returns.AuthenticatedCLIRunner = boshCLI

			err := manager.Update(incomingState)
			Expect(err).NotTo(HaveOccurred())

			Expect(opsGenerator.GenerateVarsCall.Receives.State).To(Equal(incomingState))

			Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join("some-vars-dir", "cloud-config-vars.yml")))
			Expect(fileIO.WriteFileCall.Receives[0].Contents).To(MatchYAML([]byte("some-vars")))
			Expect(err).NotTo(HaveOccurred())

			Expect(configUpdater.InitializeAuthenticatedCLICall.CallCount).To(Equal(1))
			Expect(configUpdater.InitializeAuthenticatedCLICall.Receives.State).To(Equal(incomingState))

			Expect(configUpdater.UpdateCloudConfigCall.CallCount).To(Equal(1))
			Expect(configUpdater.UpdateCloudConfigCall.Receives.AuthenticatedCLIRunner).To(Equal(boshCLI))
			Expect(configUpdater.UpdateCloudConfigCall.Receives.Filepath).To(Equal(filepath.Join(cloudConfigDir, "cloud-config.yml")))
			Expect(configUpdater.UpdateCloudConfigCall.Receives.VarsFilepath).To(Equal(filepath.Join(varsDir, "cloud-config-vars.yml")))
			Expect(configUpdater.UpdateCloudConfigCall.Receives.OpsFilepaths).To(Equal([]string{
				filepath.Join(cloudConfigDir, "ops.yml"),
				filepath.Join(cloudConfigDir, "shenanigans-ops.yml"),
			}))

		})

		Context("failure cases", func() {
			Context("when the config updater fails to initialize the authenticated bosh cli", func() {
				BeforeEach(func() {
					configUpdater.InitializeAuthenticatedCLICall.Returns.Error = errors.New("naval")
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to initialize authenticated bosh cli: naval"))
				})
			})

			Context("when getting the vars dir fails", func() {
				It("returns an error", func() {
					dirProvider.GetVarsDirCall.Returns.Directory = ""
					dirProvider.GetVarsDirCall.Returns.Error = errors.New("avocado")

					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("could not find vars directory: avocado"))
				})
			})

			Context("when ops generator fails to generate vars", func() {
				BeforeEach(func() {
					opsGenerator.GenerateVarsCall.Returns.Error = errors.New("raspberry")
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to generate cloud config vars: raspberry"))
				})
			})

			Context("when write file fails to write the vars file", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{
						Error: errors.New("apple"),
					}}
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to write cloud config vars: apple"))
				})
			})

			Context("when getting the cloud-config dir fails", func() {
				It("returns an error", func() {
					dirProvider.GetCloudConfigDirCall.Returns.Directory = ""
					dirProvider.GetCloudConfigDirCall.Returns.Error = errors.New("lime")

					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("could not find cloud-config directory: lime"))
				})
			})

			Context("when reading the cloud-config directory fails", func() {
				BeforeEach(func() {
					fileIO.ReadDirCall.Returns.Error = errors.New("aubergine")
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to read the cloud-config directory: aubergine"))
				})
			})

			Context("when the config updater fails to update the cloud config", func() {
				BeforeEach(func() {
					configUpdater.UpdateCloudConfigCall.Returns.Error = errors.New("mandarin")
				})

				It("returns an error", func() {
					err := manager.Update(storage.State{})
					Expect(err).To(MatchError("failed to update cloud-config: mandarin"))
				})
			})
		})
	})
})
