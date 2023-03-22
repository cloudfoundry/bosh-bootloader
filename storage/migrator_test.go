package storage_test

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Migrator", func() {
	var (
		migrator          storage.Migrator
		store             *fakes.StateStore
		fileIO            *fakes.FileIO
		incomingState     storage.State
		stateDir          string
		varsDir           string
		terraformDir      string
		oldBblDir         string
		oldCloudConfigDir string
		cloudConfigDir    string
	)

	BeforeEach(func() {
		store = &fakes.StateStore{}
		fileIO = &fakes.FileIO{}
		migrator = storage.NewMigrator(store, fileIO)

		var err error
		stateDir, err = os.MkdirTemp("", "")
		Expect(err).NotTo(HaveOccurred())

		cloudConfigDir = filepath.Join(stateDir, "cloud-config")
		err = os.Mkdir(cloudConfigDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		varsDir = filepath.Join(stateDir, "vars")
		err = os.Mkdir(varsDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		terraformDir = filepath.Join(stateDir, "terraform")
		err = os.Mkdir(terraformDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		oldBblDir = filepath.Join(stateDir, ".bbl")
		err = os.Mkdir(oldBblDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		oldCloudConfigDir = filepath.Join(stateDir, ".bbl", "cloudconfig")
		err = os.Mkdir(oldCloudConfigDir, os.ModePerm)
		Expect(err).NotTo(HaveOccurred())

		store.GetCloudConfigDirCall.Returns.Directory = cloudConfigDir
		store.GetVarsDirCall.Returns.Directory = varsDir
		store.GetOldBblDirCall.Returns.Directory = oldBblDir
	})

	Describe("MigrateJumpboxVars", func() {
		Context("when the state has populated jumpbox variables", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					EnvID: "some-env-id",
					Jumpbox: storage.Jumpbox{
						URL:       "10.0.0.5:25555",
						Variables: "some-jumpbox-vars",
					},
				}
				fileIO.StatCall.Returns.Error = errors.New("nope")
			})

			It("copies the jumpbox state to the jumpbox-vars-store.yml file", func() {
				_, err := migrator.MigrateJumpboxVars(incomingState, varsDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join(varsDir, "jumpbox-vars-store.yml")))
				Expect(string(fileIO.WriteFileCall.Receives[0].Contents)).To(Equal("some-jumpbox-vars"))
			})

			Context("when the director variables file cannot be written", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("persimmon")}}
				})

				It("returns an error", func() {
					_, err := migrator.MigrateJumpboxVars(incomingState, varsDir)
					Expect(err).To(MatchError(ContainSubstring("migrating jumpbox variables: persimmon")))
				})
			})
		})

		Context("when jumpbox variables are in jumpbox-variables.yml", func() {
			BeforeEach(func() {
				fileIO.ReadFileCall.Returns.Contents = []byte("some-jumpbox-vars")

				incomingState = storage.State{EnvID: "some-env-id"}
			})

			It("moves jumpbox-variables.yml to jumpbox-vars-store.yml", func() {
				_, err := migrator.MigrateJumpboxVars(incomingState, varsDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join(varsDir, "jumpbox-vars-store.yml")))
				Expect(string(fileIO.WriteFileCall.Receives[0].Contents)).To(Equal("some-jumpbox-vars"))
			})

			Context("when the jumpbox legacy vars-store file cannot be read", func() {
				BeforeEach(func() {
					fileIO.ReadFileCall.Returns.Error = errors.New("pomegranate")
				})

				It("returns an error", func() {
					_, err := migrator.MigrateJumpboxVars(incomingState, varsDir)
					Expect(err).To(MatchError(ContainSubstring("reading legacy jumpbox vars store: pomegranate")))
				})
			})
		})
	})

	Describe("MigrateDirectorVars", func() {
		Context("when the state has populated BOSH variables", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					EnvID: "some-env-id",
					BOSH: storage.BOSH{
						DirectorAddress: "10.0.0.6",
						Variables:       "some-director-vars",
					},
				}
				fileIO.StatCall.Returns.Error = errors.New("nope")
			})

			It("copies the BOSH state to the director-vars-store.yml file", func() {
				_, err := migrator.MigrateDirectorVars(incomingState, varsDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(fileIO.WriteFileCall.Receives[0].Contents)).To(Equal("some-director-vars"))
				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join(varsDir, "director-vars-store.yml")))
			})

			Context("when the director variables file cannot be written", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("watermelon")}}
				})

				It("returns an error", func() {
					_, err := migrator.MigrateDirectorVars(incomingState, varsDir)
					Expect(err).To(MatchError(ContainSubstring("migrating director variables: watermelon")))
				})
			})
		})

		Context("when BOSH variables are in director-variables.yml", func() {
			BeforeEach(func() {
				fileIO.ReadFileCall.Returns.Contents = []byte("some-director-vars")

				incomingState = storage.State{EnvID: "some-env"}
			})

			It("moves director-variables.yml to director-vars-store.yml", func() {
				_, err := migrator.MigrateDirectorVars(storage.State{EnvID: "some-env"}, varsDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join(varsDir, "director-vars-store.yml")))
				Expect(fileIO.WriteFileCall.Receives[0].Contents).To(Equal([]byte("some-director-vars")))
			})

			Context("failure cases", func() {
				Context("when the director legacy vars-store file cannot be read", func() {
					BeforeEach(func() {
						fileIO.ReadFileCall.Returns.Error = errors.New("pomelo")
					})

					It("returns an error", func() {
						_, err := migrator.MigrateDirectorVars(incomingState, varsDir)
						Expect(err).To(MatchError(ContainSubstring("reading legacy director vars store: pomelo")))
					})
				})
			})
		})
	})

	Describe("MigrateDirectorVarsFile", func() {
		Context("when the state has a director-deployment-vars.yml file", func() {
			It("migrates the file to director-vars-file.yml", func() {
				err := migrator.MigrateDirectorVarsFile(varsDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.RenameCall.Receives.Oldpath).To(Equal(filepath.Join(varsDir, "director-deployment-vars.yml")))
				Expect(fileIO.RenameCall.Receives.Newpath).To(Equal(filepath.Join(varsDir, "director-vars-file.yml")))
			})
		})
		Context("when renaming the director-deployment-vars.yml file fails", func() {
			BeforeEach(func() {
				fileIO.RenameCall.Returns.Error = errors.New("pumpkins aren't a fruit")
			})

			It("returns an error", func() {
				err := migrator.MigrateDirectorVarsFile(varsDir)
				Expect(err).To(MatchError(ContainSubstring("pumpkins")))
			})
		})
	})

	Describe("MigrateJumpboxVarsFile", func() {
		Context("when the state has a jumpbox-deployment-vars.yml file", func() {
			It("migrates the file to jumpbox-vars-file.yml", func() {
				err := migrator.MigrateJumpboxVarsFile(varsDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.RenameCall.Receives.Oldpath).To(Equal(filepath.Join(varsDir, "jumpbox-deployment-vars.yml")))
				Expect(fileIO.RenameCall.Receives.Newpath).To(Equal(filepath.Join(varsDir, "jumpbox-vars-file.yml")))
			})
		})
		Context("when renaming the jumpbox-deployment-vars.yml file fails", func() {
			BeforeEach(func() {
				fileIO.RenameCall.Returns.Error = errors.New("pumpkins aren't a fruit")
			})

			It("returns an error", func() {
				err := migrator.MigrateJumpboxVarsFile(varsDir)
				Expect(err).To(MatchError(ContainSubstring("pumpkins")))
			})
		})
	})

	Describe("MigrateTFVars", func() {
		Context("when the state has bbl-provided tfvars in the terraform.tfvars file", func() {
			It("migrates terraform.tfvars to bbl.tfvars", func() {
				err := migrator.MigrateTerraformVars(varsDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.RenameCall.Receives.Oldpath).To(Equal(filepath.Join(varsDir, "terraform.tfvars")))
				Expect(fileIO.RenameCall.Receives.Newpath).To(Equal(filepath.Join(varsDir, "bbl.tfvars")))
			})

			Context("when renaming the tfvars file fails", func() {
				BeforeEach(func() {
					fileIO.RenameCall.Returns.Error = errors.New("potatoes aren't a fruit")
				})

				It("returns an error", func() {
					err := migrator.MigrateTerraformVars(varsDir)
					Expect(err).To(MatchError(ContainSubstring("potatoes")))
				})
			})
		})
	})

	Describe("MigrateCloudConfigDir", func() {
		Context("when the state has a populated .bbl directory", func() {
			BeforeEach(func() {
				fileIO.ReadDirCall.Returns.FileInfos = []os.FileInfo{
					fakes.FileInfo{
						FileName: "some-config-file",
					},
				}
				fileIO.ReadFileCall.Returns.Contents = []byte("some-cloud-config")
			})

			It("moves the cloud-config directory to the top level", func() {
				err := migrator.MigrateCloudConfigDir(oldBblDir, cloudConfigDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join(cloudConfigDir, "some-config-file")))
				Expect(string(fileIO.WriteFileCall.Receives[0].Contents)).To(Equal("some-cloud-config"))
				Expect(fileIO.RemoveAllCall.Receives[0].Path).To(Equal(oldBblDir))
			})

			Context("when the contents of the old .bbl dir cannot be read", func() {
				BeforeEach(func() {
					fileIO.ReadDirCall.Returns.Error = errors.New("kiwano")
				})

				It("returns an error", func() {
					err := migrator.MigrateCloudConfigDir(oldBblDir, cloudConfigDir)
					Expect(err).To(MatchError(ContainSubstring("reading legacy .bbl dir contents: ")))
				})
			})

			Context("when the reading the old cloud config file fails", func() {
				BeforeEach(func() {
					fileIO.ReadFileCall.Returns.Error = errors.New("durian")
				})

				It("returns an error", func() {
					err := migrator.MigrateCloudConfigDir(oldBblDir, cloudConfigDir)
					Expect(err).To(MatchError(ContainSubstring("reading")))
				})
			})

			Context("when renaming a cloud-config file fails", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("tamarillo")}}
				})

				It("returns an error", func() {
					err := migrator.MigrateCloudConfigDir(oldBblDir, cloudConfigDir)
					Expect(err).To(MatchError(ContainSubstring("migrating")))
				})
			})

			Context("when removing the old .bbl dir fails", func() {
				BeforeEach(func() {
					fileIO.RemoveAllCall.Returns = []fakes.RemoveAllReturn{{Error: errors.New("feijoa")}}
				})

				It("returns an error", func() {
					err := migrator.MigrateCloudConfigDir(oldBblDir, cloudConfigDir)
					Expect(err).To(MatchError(ContainSubstring("removing legacy .bbl dir: ")))
				})
			})
		})
	})

	Describe("MigrateJumpboxState", func() {
		Context("when the state has a populated jumpbox state", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					EnvID: "some-env-id",
					Jumpbox: storage.Jumpbox{
						URL: "10.0.0.5",
						State: map[string]interface{}{
							"some-jumpbox-key": "some-jumpbox-value",
						},
					},
				}
			})

			It("copies the jumpbox state to the jumpbox-state.json vars file", func() {
				state, err := migrator.MigrateJumpboxState(incomingState, varsDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.WriteFileCall.Receives[0].Contents).To(MatchJSON(`{"some-jumpbox-key": "some-jumpbox-value"}`))
				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join(varsDir, "jumpbox-state.json")))
				Expect(state.Jumpbox.State).To(BeNil())
			})

			Context("when the jumpbox state file cannot be written", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("pitaya")}}
				})

				It("returns an error", func() {
					_, err := migrator.MigrateJumpboxState(incomingState, varsDir)
					Expect(err).To(MatchError(ContainSubstring("migrating jumpbox state: ")))
				})
			})

			Context("when the jumpbox state file cannot be written", func() {
				BeforeEach(func() {
					incomingState.Jumpbox.State["invalid-key"] = func() string { return "invalid" }
				})

				It("returns an error", func() {
					_, err := migrator.MigrateJumpboxState(incomingState, varsDir)
					Expect(err).To(MatchError(ContainSubstring("marshalling jumpbox state: ")))
				})
			})
		})
	})

	Describe("MigrateDirectorState", func() {
		Context("when the state has a populated BOSH state", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					EnvID: "some-env-id",
					BOSH: storage.BOSH{
						DirectorAddress: "10.0.0.6",
						State: map[string]interface{}{
							"some-bosh-key": "some-bosh-value",
						},
					},
				}
			})

			It("copies the BOSH state to the bosh-state.json vars file", func() {
				state, err := migrator.MigrateDirectorState(incomingState, varsDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.WriteFileCall.Receives[0].Contents).To(MatchJSON(`{"some-bosh-key": "some-bosh-value"}`))
				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join(varsDir, "bosh-state.json")))
				Expect(state.BOSH.State).To(BeNil())
			})

			Context("when the bosh state file cannot be written", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("jackfruit")}}
				})

				It("returns an error", func() {
					_, err := migrator.MigrateDirectorState(incomingState, varsDir)
					Expect(err).To(MatchError(ContainSubstring("migrating bosh state: ")))
				})
			})

			Context("when the bosh state file cannot be marshalled", func() {
				BeforeEach(func() {
					incomingState.BOSH.State["invalid-key"] = func() string { return "invalid" }
				})

				It("returns an error", func() {
					_, err := migrator.MigrateDirectorState(incomingState, varsDir)
					Expect(err).To(MatchError(ContainSubstring("marshalling bosh state: ")))
				})
			})
		})
	})

	Describe("MigrateTerraformState", func() {
		Context("when the state has a populated TFState", func() {
			BeforeEach(func() {
				incomingState = storage.State{
					EnvID:   "some-env-id",
					TFState: "some-tf-state",
				}
			})

			It("writes the TFState to the tfstate file", func() {
				outgoingState, err := migrator.MigrateTerraformState(incomingState, varsDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(outgoingState.TFState).To(Equal(""))

				Expect(string(fileIO.WriteFileCall.Receives[0].Contents)).To(Equal("some-tf-state"))
				Expect(fileIO.WriteFileCall.Receives[0].Filename).To(Equal(filepath.Join(varsDir, "terraform.tfstate")))
			})

			Context("when the tfstate file cannot be written", func() {
				BeforeEach(func() {
					fileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("cherimoya")}}
				})

				It("returns an error", func() {
					_, err := migrator.MigrateTerraformState(incomingState, varsDir)
					Expect(err).To(MatchError(ContainSubstring("migrating terraform state: ")))
				})
			})
		})
	})

	Describe("MigrateTerraformTemplate", func() {
		Context("when a template.tf file exists", func() {
			It("writes the TFState to the tfstate file", func() {
				err := migrator.MigrateTerraformTemplate(terraformDir)
				Expect(err).NotTo(HaveOccurred())

				Expect(fileIO.RenameCall.Receives.Oldpath).To(Equal(filepath.Join(terraformDir, "template.tf")))
				Expect(fileIO.RenameCall.Receives.Newpath).To(Equal(filepath.Join(terraformDir, "bbl-template.tf")))
			})
		})

		Context("when the template file cannot be renamed", func() {
			BeforeEach(func() {
				fileIO.RenameCall.Returns.Error = errors.New("apple")
			})

			It("returns an error", func() {
				err := migrator.MigrateTerraformTemplate(terraformDir)
				Expect(err).To(MatchError(ContainSubstring("migrating terraform template: ")))
			})
		})
	})

	Describe("Migrate", func() {
		Context("when the state is empty", func() {
			It("returns the state without changing it", func() {
				outgoingState, err := migrator.Migrate(storage.State{})
				Expect(err).NotTo(HaveOccurred())
				Expect(outgoingState).To(Equal(storage.State{}))
				Expect(store.SetCall.CallCount).To(Equal(0))
			})
		})

		Context("when the state is already migrated", func() {
			BeforeEach(func() {
				incomingState = storage.State{EnvID: "some-env-id"}
			})

			Context("when the vars dir cannot be retrieved", func() {
				BeforeEach(func() {
					store.GetVarsDirCall.Returns.Error = errors.New("potato")
				})

				It("returns an error", func() {
					_, err := migrator.Migrate(incomingState)
					Expect(err).To(MatchError("migrating state: potato"))
				})
			})

			Context("when the state cannot be saved", func() {
				BeforeEach(func() {
					store.SetCall.Returns = []fakes.SetCallReturn{
						{
							Error: errors.New("tomato"),
						},
					}
				})

				It("returns an error", func() {
					_, err := migrator.Migrate(incomingState)
					Expect(err).To(MatchError("saving migrated state: tomato"))
				})
			})
		})
	})
})
