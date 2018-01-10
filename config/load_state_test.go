package config_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/config"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadState", func() {
	var (
		fakeLogger         *fakes.Logger
		fakeStateBootstrap *fakes.StateBootstrap
		fakeStateMigrator  *fakes.StateMigrator
		fakeFileIO         *fakes.FileIO
		c                  config.Config
	)

	BeforeEach(func() {
		fakeLogger = &fakes.Logger{}
		fakeStateBootstrap = &fakes.StateBootstrap{}
		fakeStateMigrator = &fakes.StateMigrator{}
		fakeFileIO = &fakes.FileIO{}
		c = config.NewConfig(fakeStateBootstrap, fakeStateMigrator, fakeLogger, fakeFileIO)
		os.Clearenv()
	})

	Describe("Bootstrap", func() {
		Describe("help and version", func() {
			Context("when no commands are passed", func() {
				It("sets the help command", func() {
					appConfig, err := c.Bootstrap([]string{"bbl"})
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("help"))
					Expect(fakeStateBootstrap.GetStateCall.CallCount).To(Equal(0))
					Expect(fakeStateMigrator.MigrateCall.CallCount).To(Equal(0))
				})
			})

			Context("when --help is passed as a flag", func() {
				It("sets the help command", func() {
					appConfig, err := c.Bootstrap([]string{"bbl", "--help"})
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("help"))
					Expect(appConfig.ShowCommandHelp).To(BeFalse())
					Expect(fakeStateBootstrap.GetStateCall.CallCount).To(Equal(0))
					Expect(fakeStateMigrator.MigrateCall.CallCount).To(Equal(0))
				})
			})

			Context("when version is passed as a flag", func() {
				It("sets the version command", func() {
					appConfig, err := c.Bootstrap([]string{"bbl", "--version"})
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("version"))
					Expect(fakeStateBootstrap.GetStateCall.CallCount).To(Equal(0))
					Expect(fakeStateMigrator.MigrateCall.CallCount).To(Equal(0))
				})
			})

			DescribeTable("subcommand help for help and version",
				func(args []string, expectedCommand string) {
					appConfig, err := c.Bootstrap(args)
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal(expectedCommand))
					Expect(appConfig.ShowCommandHelp).To(BeTrue())
					Expect(fakeStateMigrator.MigrateCall.CallCount).To(Equal(0))
				},
				Entry("bbl help help", []string{"bbl", "help", "help"}, "help"),
				Entry("bbl help version", []string{"bbl", "help", "version"}, "version"),
				Entry("bbl version --help", []string{"bbl", "version", "--help"}, "version"),
				Entry("bbl help rotate", []string{"bbl", "help", "rotate"}, "rotate"),
			)
		})

		Describe("global flags", func() {
			var fullStateDirPath string
			BeforeEach(func() {
				workingDir, err := os.Getwd()
				Expect(err).NotTo(HaveOccurred())

				fullStateDirPath = filepath.Join(workingDir, "some-state-dir")
			})

			It("returns global flags", func() {
				args := []string{
					"bbl", "up",
					"--debug",
					"--state-dir", "some-state-dir",
				}

				appConfig, err := c.Bootstrap(args)
				Expect(err).NotTo(HaveOccurred())

				Expect(appConfig.Command).To(Equal("up"))
				Expect(appConfig.Global.Debug).To(BeTrue())
				Expect(appConfig.Global.StateDir).To(Equal(fullStateDirPath))
			})

			Context("when --help is passed in after a command", func() {
				It("returns command help", func() {
					args := []string{
						"bbl", "up",
						"--help",
					}

					appConfig, err := c.Bootstrap(args)
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("up"))
					Expect(appConfig.ShowCommandHelp).To(BeTrue())
				})
			})

			Context("when help is passed in before a command", func() {
				It("returns command help", func() {
					args := []string{"bbl", "help", "up"}

					appConfig, err := c.Bootstrap(args)
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("up"))
					Expect(appConfig.ShowCommandHelp).To(BeTrue())
				})
			})

			Context("when debug flag is passed in through environment variable", func() {
				BeforeEach(func() {
					os.Setenv("BBL_DEBUG", "true")
				})

				AfterEach(func() {
					os.Unsetenv("BBL_DEBUG")
				})

				It("returns global flags", func() {
					appConfig, err := c.Bootstrap([]string{"bbl", "up"})
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Global.Debug).To(BeTrue())
				})
			})

			Context("when state dir flag is passed in through environment variable", func() {
				BeforeEach(func() {
					os.Setenv("BBL_STATE_DIRECTORY", "/path/to/state")
				})

				AfterEach(func() {
					os.Unsetenv("BBL_STATE_DIRECTORY")
				})

				It("returns global flags", func() {
					appConfig, err := c.Bootstrap([]string{"bbl", "up"})
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Global.StateDir).To(Equal("/path/to/state"))
				})
			})
		})

		Describe("reading a previous state file", func() {
			var (
				gotState      storage.State
				migratedState storage.State
			)
			BeforeEach(func() {
				gotState = storage.State{
					IAAS:    "aws",
					EnvID:   "some-env-id",
					TFState: "some-tf-state",
				}
				migratedState = storage.State{
					IAAS:  "aws",
					EnvID: "some-env-id",
				}
				fakeStateBootstrap.GetStateCall.Returns.State = gotState
				fakeStateMigrator.MigrateCall.Returns.State = migratedState
			})

			It("returns the existing state", func() {
				appConfig, err := c.Bootstrap([]string{
					"bbl",
					"rotate",
				})
				Expect(err).NotTo(HaveOccurred())

				By("migrating the existing state file", func() {
					Expect(fakeStateBootstrap.GetStateCall.CallCount).To(Equal(1))
					Expect(fakeStateMigrator.MigrateCall.CallCount).To(Equal(1))
					Expect(fakeStateMigrator.MigrateCall.Receives.State).To(Equal(gotState))
				})

				Expect(appConfig.State).To(Equal(migratedState))
			})

			It("uses the working directory", func() {
				appConfig, err := c.Bootstrap([]string{
					"bbl",
					"rotate",
				})
				Expect(err).NotTo(HaveOccurred())

				workingDir, err := os.Getwd()
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStateBootstrap.GetStateCall.Receives.Dir).To(Equal(workingDir))
				Expect(appConfig.Global.StateDir).To(Equal(workingDir))
			})

			Context("when state dir is specified as a relative path", func() {
				var expectedDir string
				BeforeEach(func() {
					workingDir, err := os.Getwd()
					Expect(err).NotTo(HaveOccurred())

					expectedDir = filepath.Join(workingDir, "some-state-dir")
				})

				It("uses the absolute path of the state dir", func() {
					appConfig, err := c.Bootstrap([]string{
						"bbl",
						"rotate",
						"--state-dir", "some-state-dir",
					})
					Expect(err).NotTo(HaveOccurred())

					By("loading state from the specified state dir", func() {
						Expect(fakeStateBootstrap.GetStateCall.Receives.Dir).To(Equal(expectedDir))
					})

					By("setting the specified state dir on global config", func() {
						Expect(appConfig.Global.StateDir).To(Equal(expectedDir))
					})
				})
			})

			Context("when state dir is specified as an absolute path", func() {
				var stateDir string
				BeforeEach(func() {
					var err error
					stateDir, err = ioutil.TempDir("", "my-state-dir-")
					Expect(err).NotTo(HaveOccurred())
				})

				It("does not modify the path of the state dir", func() {
					appConfig, err := c.Bootstrap([]string{
						"bbl",
						"rotate",
						"--state-dir", stateDir,
					})
					Expect(err).NotTo(HaveOccurred())

					By("loading state from the specified state dir", func() {
						Expect(fakeStateBootstrap.GetStateCall.Receives.Dir).To(Equal(stateDir))
					})

					By("setting the specified state dir on global config", func() {
						Expect(appConfig.Global.StateDir).To(Equal(stateDir))
					})
				})
			})

			Context("when invalid state dir is passed in", func() {
				BeforeEach(func() {
					fakeStateBootstrap.GetStateCall.Returns.Error = errors.New("some state dir error")
					os.Clearenv()
				})

				It("returns an error", func() {
					_, err := c.Bootstrap([]string{
						"bbl",
						"rotate",
						"--state-dir", "/this/will/not/work",
					})

					Expect(err).To(MatchError("some state dir error"))
				})
			})

			Context("when migrating the state fails", func() {
				BeforeEach(func() {
					fakeStateMigrator.MigrateCall.Returns.Error = errors.New("coconut")
				})
				It("returns an error", func() {
					_, err := c.Bootstrap([]string{
						"bbl",
						"rotate",
						"--state-dir", "some-state-dir",
					})
					Expect(err).To(MatchError("coconut"))
				})
			})

			Context("when state-dir flag is passed without an argument", func() {
				It("returns an error", func() {
					_, err := c.Bootstrap([]string{
						"bbl",
						"rotate",
						"--state-dir",
						"--help",
					})

					Expect(err).To(MatchError("expected argument for flag `-s, --state-dir', but got option `--help'"))
				})
			})
		})

		Context("using VSphere", func() {
			Context("when a previous state does not exist", func() {
				Context("when configuration is passed in by flag", func() {
					var args []string

					BeforeEach(func() {
						args = []string{
							"bbl",
							"--iaas", "vsphere",
							"--vsphere-vcenter-user", "user",
							"--vsphere-vcenter-password", "password",
							"--vsphere-vcenter-ip", "ip",
							"--vsphere-vcenter-dc", "dc",
							"--vsphere-vcenter-cluster", "cluster",
							"--vsphere-vcenter-rp", "rp",
							"--vsphere-network", "network",
							"--vsphere-vcenter-ds", "ds",
							"--vsphere-subnet", "subnet",
							"up",
							"--name", "some-env-id",
						}
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(args)
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State

						Expect(state.IAAS).To(Equal("vsphere"))
						Expect(state.VSphere.VCenterUser).To(Equal("user"))
						Expect(state.VSphere.VCenterPassword).To(Equal("password"))
						Expect(state.VSphere.VCenterIP).To(Equal("ip"))
						Expect(state.VSphere.VCenterDC).To(Equal("dc"))
						Expect(state.VSphere.Cluster).To(Equal("cluster"))
						Expect(state.VSphere.VCenterRP).To(Equal("rp"))
						Expect(state.VSphere.Network).To(Equal("network"))
						Expect(state.VSphere.VCenterDS).To(Equal("ds"))
						Expect(state.VSphere.Subnet).To(Equal("subnet"))
					})

					It("returns the remaining arguments", func() {
						appConfig, err := c.Bootstrap(args)
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.SubcommandFlags).To(Equal(application.StringSlice{"--name", "some-env-id"}))
					})
				})

				Context("when configuration is passed in by env vars", func() {
					var args []string

					BeforeEach(func() {
						args = []string{"bbl", "up"}

						os.Setenv("BBL_IAAS", "vsphere")
						os.Setenv("BBL_VSPHERE_VCENTER_USER", "user")
						os.Setenv("BBL_VSPHERE_VCENTER_PASSWORD", "password")
						os.Setenv("BBL_VSPHERE_VCENTER_IP", "ip")
						os.Setenv("BBL_VSPHERE_VCENTER_DC", "dc")
						os.Setenv("BBL_VSPHERE_VCENTER_CLUSTER", "cluster")
						os.Setenv("BBL_VSPHERE_VCENTER_RP", "rp")
						os.Setenv("BBL_VSPHERE_NETWORK", "network")
						os.Setenv("BBL_VSPHERE_VCENTER_DS", "ds")
						os.Setenv("BBL_VSPHERE_SUBNET", "subnet")
					})

					AfterEach(func() {
						os.Unsetenv("BBL_IAAS")
						os.Unsetenv("BBL_VSPHERE_VCENTER_USER")
						os.Unsetenv("BBL_VSPHERE_VCENTER_PASSWORD")
						os.Unsetenv("BBL_VSPHERE_VCENTER_IP")
						os.Unsetenv("BBL_VSPHERE_VCENTER_DC")
						os.Unsetenv("BBL_VSPHERE_VCENTER_CLUSTER")
						os.Unsetenv("BBL_VSPHERE_VCENTER_RP")
						os.Unsetenv("BBL_VSPHERE_NETWORK")
						os.Unsetenv("BBL_VSPHERE_VCENTER_DS")
						os.Unsetenv("BBL_VSPHERE_SUBNET")
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(args)
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State
						Expect(state.IAAS).To(Equal("vsphere"))
						Expect(state.VSphere.VCenterUser).To(Equal("user"))
						Expect(state.VSphere.VCenterPassword).To(Equal("password"))
						Expect(state.VSphere.VCenterIP).To(Equal("ip"))
						Expect(state.VSphere.VCenterDC).To(Equal("dc"))
						Expect(state.VSphere.Cluster).To(Equal("cluster"))
						Expect(state.VSphere.VCenterRP).To(Equal("rp"))
						Expect(state.VSphere.Network).To(Equal("network"))
						Expect(state.VSphere.VCenterDS).To(Equal("ds"))
						Expect(state.VSphere.Subnet).To(Equal("subnet"))
					})

					It("returns the command", func() {
						appConfig, err := c.Bootstrap(args)
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
					})
				})
			})

			Context("when a previous state exists", func() {
				BeforeEach(func() {
					fakeStateMigrator.MigrateCall.Returns.State = storage.State{
						IAAS: "vsphere",
						VSphere: storage.VSphere{
							VCenterUser:     "user",
							VCenterPassword: "password",
							VCenterIP:       "ip",
							VCenterDC:       "dc",
							Cluster:         "cluster",
							VCenterRP:       "rp",
							Network:         "network",
							VCenterDS:       "ds",
							Subnet:          "subnet",
						},
						EnvID: "some-env-id",
					}
				})

				Context("when valid matching configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						appConfig, err := c.Bootstrap([]string{
							"bbl", "up",
							"--iaas", "vsphere",
							"--vsphere-vcenter-user", "user",
							"--vsphere-vcenter-password", "password",
							"--vsphere-vcenter-ip", "ip",
							"--vsphere-vcenter-dc", "dc",
							"--vsphere-vcenter-cluster", "cluster",
							"--vsphere-vcenter-rp", "rp",
							"--vsphere-network", "network",
							"--vsphere-vcenter-ds", "ds",
							"--vsphere-subnet", "subnet",
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.State.EnvID).To(Equal("some-env-id"))
					})
				})

				DescribeTable("when non-matching configuration is passed in",
					func(args []string, expected string) {
						_, err := c.Bootstrap(args)

						Expect(err).To(MatchError(expected))
					},
					Entry("returns an error for non-matching IAAS", []string{"bbl", "up", "--iaas", "gcp"},
						"The iaas type cannot be changed for an existing environment. The current iaas type is vsphere."),
				)
			})
		})

		Context("using AWS", func() {
			Context("when a previous state does not exist", func() {
				Context("when configuration is passed in by flag", func() {
					var args []string

					BeforeEach(func() {
						args = []string{
							"bbl",
							"--iaas", "aws",
							"--aws-access-key-id", "some-access-key",
							"--aws-secret-access-key", "some-secret-key",
							"--aws-region", "some-region",
							"up",
							"--name", "some-env-id",
						}
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(args)
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State

						Expect(state.IAAS).To(Equal("aws"))
						Expect(state.AWS.AccessKeyID).To(Equal("some-access-key"))
						Expect(state.AWS.SecretAccessKey).To(Equal("some-secret-key"))
						Expect(state.AWS.Region).To(Equal("some-region"))
					})

					It("returns the remaining arguments", func() {
						appConfig, err := c.Bootstrap(args)
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.SubcommandFlags).To(Equal(application.StringSlice{"--name", "some-env-id"}))
					})
				})

				Context("when configuration is passed in by env vars", func() {
					var args []string

					BeforeEach(func() {
						args = []string{
							"bbl", "up",
						}

						os.Setenv("BBL_IAAS", "aws")
						os.Setenv("BBL_AWS_ACCESS_KEY_ID", "some-access-key-id")
						os.Setenv("BBL_AWS_SECRET_ACCESS_KEY", "some-secret-key")
						os.Setenv("BBL_AWS_REGION", "some-region")
					})

					AfterEach(func() {
						os.Unsetenv("BBL_IAAS")
						os.Unsetenv("BBL_AWS_ACCESS_KEY_ID")
						os.Unsetenv("BBL_AWS_SECRET_ACCESS_KEY")
						os.Unsetenv("BBL_AWS_REGION")
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(args)
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State
						Expect(state.IAAS).To(Equal("aws"))
						Expect(state.AWS.AccessKeyID).To(Equal("some-access-key-id"))
						Expect(state.AWS.SecretAccessKey).To(Equal("some-secret-key"))
						Expect(state.AWS.Region).To(Equal("some-region"))
					})

					It("returns the command", func() {
						appConfig, err := c.Bootstrap(args)
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
					})
				})
			})

			Context("when a previous state exists", func() {
				BeforeEach(func() {
					fakeStateMigrator.MigrateCall.Returns.State = storage.State{
						IAAS: "aws",
						AWS: storage.AWS{
							AccessKeyID:     "some-access-key-id",
							SecretAccessKey: "some-secret-access-key",
							Region:          "some-region",
						},
						EnvID: "some-env-id",
					}
				})

				Context("when valid matching configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						appConfig, err := c.Bootstrap([]string{
							"bbl", "up",
							"--iaas", "aws",
							"--aws-access-key-id", "some-access-key-id",
							"--aws-secret-access-key", "some-secret-access-key",
							"--aws-region", "some-region",
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.State.EnvID).To(Equal("some-env-id"))
					})
				})

				DescribeTable("when non-matching configuration is passed in",
					func(args []string, expected string) {
						_, err := c.Bootstrap(args)

						Expect(err).To(MatchError(expected))
					},
					Entry("returns an error for non-matching IAAS", []string{"bbl", "up", "--iaas", "gcp"},
						"The iaas type cannot be changed for an existing environment. The current iaas type is aws."),
					Entry("returns an error for non-matching region", []string{"bbl", "up", "--aws-region", "some-other-region"},
						"The region cannot be changed for an existing environment. The current region is some-region."),
				)
			})
		})

		Context("using GCP", func() {
			var (
				serviceAccountKeyPath string
				serviceAccountKey     string
				tempFile              *os.File
			)
			BeforeEach(func() {
				var err error
				tempFile, err = ioutil.TempFile("", "temp")
				Expect(err).NotTo(HaveOccurred())
				serviceAccountKeyPath = tempFile.Name()
				serviceAccountKey = `{"project_id": "some-project-id"}`

				fakeFileIO.TempFileCall.Returns.File = tempFile
				fakeFileIO.ReadFileCall.Returns.Contents = []byte(serviceAccountKey)
			})

			Context("when a previous state does not exist", func() {
				Context("when configuration is passed in by flag", func() {
					var args []string

					BeforeEach(func() {
						args = []string{
							"bbl", "up", "--name", "some-env-id",
							"--iaas", "gcp",
							"--gcp-service-account-key", "/path/to/service/account/key",
							"--gcp-region", "some-region",
						}
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State
						Expect(state.IAAS).To(Equal("gcp"))
						Expect(state.GCP.ServiceAccountKeyPath).To(Equal("/path/to/service/account/key"))
						Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
						Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
						Expect(state.GCP.Region).To(Equal("some-region"))
					})

					It("returns the command and its flags", func() {
						appConfig, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.SubcommandFlags).To(Equal(application.StringSlice{"--name", "some-env-id"}))
					})

					Context("when service account key is passed inline", func() {
						var args []string

						BeforeEach(func() {
							fakeFileIO.StatCall.Returns.Error = errors.New("no file found")
							args = []string{
								"bbl", "up", "--name", "some-env-id",
								"--iaas", "gcp",
								"--gcp-service-account-key", serviceAccountKey,
								"--gcp-region", "some-region",
							}
						})

						It("returns a state object containing a path to the service account key", func() {
							appConfig, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())
							Expect(appConfig.State.GCP.ProjectID).To(Equal("some-project-id"))
							Expect(fakeFileIO.WriteFileCall.Receives.Filename).To(Equal(tempFile.Name()))
							Expect(fakeFileIO.WriteFileCall.Receives.Contents).To(Equal([]byte(serviceAccountKey)))
						})
					})

					Context("when service account key is invalid", func() {
						var args []string

						Context("when service account key file is missing", func() {
							BeforeEach(func() {
								args = []string{
									"bbl",
									"up",
									"--iaas", "gcp",
									"--gcp-service-account-key", "/this/file/isn't/real",
									"--gcp-region", "some-region",
								}
								fakeFileIO.StatCall.Returns.Error = errors.New("no file found")
							})

							It("returns an error", func() {
								_, err := c.Bootstrap(args)

								Expect(err).To(MatchError("Unmarshalling service account key (must be valid json): invalid character '/' looking for beginning of value"))
							})
						})

						Context("when service account key is invalid json", func() {
							BeforeEach(func() {
								args = []string{
									"bbl", "up", "--name", "some-env-id",
									"--iaas", "gcp",
									"--gcp-service-account-key", "some-key",
									"--gcp-region", "some-region",
								}
								fakeFileIO.ReadFileCall.Returns.Contents = []byte("not-json")
							})

							It("returns an error", func() {
								_, err := c.Bootstrap(args)
								Expect(err).To(MatchError(ContainSubstring("Unmarshalling service account key (must be valid json):")))
							})
						})
					})

					Context("when service account key is missing project ID field", func() {
						BeforeEach(func() {
							args = []string{
								"bbl", "up", "--name", "some-env-id",
								"--iaas", "gcp",
								"--gcp-service-account-key", "some-key",
								"--gcp-region", "some-region",
							}
							fakeFileIO.ReadFileCall.Returns.Contents = []byte("{}")
						})

						It("returns an error", func() {
							_, err := c.Bootstrap(args)
							Expect(err).To(MatchError("Service account key is missing field `project_id`"))
						})
					})

					Context("when the temp file cannot be created", func() {
						BeforeEach(func() {
							args = []string{
								"bbl", "up", "--name", "some-env-id",
								"--iaas", "gcp",
								"--gcp-service-account-key", "some-key",
								"--gcp-region", "some-region",
							}
							fakeFileIO.StatCall.Returns.Error = errors.New("no file found")
							fakeFileIO.TempFileCall.Returns.Error = errors.New("banana")
						})

						It("returns an error", func() {
							_, err := c.Bootstrap(args)
							Expect(err).To(MatchError("Creating temp file for credentials: banana"))
						})
					})

					Context("when writing the key to the file fails", func() {
						BeforeEach(func() {
							args = []string{
								"bbl", "up", "--name", "some-env-id",
								"--iaas", "gcp",
								"--gcp-service-account-key", "some-key",
								"--gcp-region", "some-region",
							}
							fakeFileIO.StatCall.Returns.Error = errors.New("no file found")
							fakeFileIO.WriteFileCall.Returns.Error = errors.New("coconut")
						})

						It("returns an error", func() {
							_, err := c.Bootstrap(args)
							Expect(err).To(MatchError("Writing credentials to temp file: coconut"))
						})
					})
				})

				Context("when configuration is passed in by env vars", func() {
					var args []string

					BeforeEach(func() {
						args = []string{"bbl", "up"}

						os.Setenv("BBL_IAAS", "gcp")
						os.Setenv("BBL_GCP_SERVICE_ACCOUNT_KEY", serviceAccountKey)
						os.Setenv("BBL_GCP_REGION", "some-region")

						fakeFileIO.StatCall.Returns.Error = errors.New("no file found")
					})

					It("returns a state containing configuration", func() {
						appConfig, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State

						Expect(state.IAAS).To(Equal("gcp"))
						Expect(state.GCP.ServiceAccountKeyPath).To(Equal(tempFile.Name()))
						Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
						Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
						Expect(state.GCP.Region).To(Equal("some-region"))
					})

					It("returns the remaining arguments", func() {
						appConfig, err := c.Bootstrap(args)
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.SubcommandFlags).To(Equal(application.StringSlice{}))
					})
				})
			})

			Context("when a previous state exists", func() {
				var existingState storage.State
				BeforeEach(func() {
					existingState = storage.State{
						IAAS: "gcp",
						GCP: storage.GCP{
							ProjectID: "some-project-id",
							Zone:      "some-zone",
							Region:    "some-region",
						},
						EnvID: "some-env-id",
					}
					fakeStateMigrator.MigrateCall.Returns.State = existingState
					fakeFileIO.StatCall.Returns.Error = errors.New("no file found")
				})

				Context("when valid matching configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						appConfig, err := c.Bootstrap([]string{
							"bbl", "up",
							"--iaas", "gcp",
							"--gcp-service-account-key", serviceAccountKey,
							"--gcp-region", "some-region",
						})
						Expect(err).NotTo(HaveOccurred())

						appConfig.State.GCP.ServiceAccountKey = ""     // this isn't written to disk
						appConfig.State.GCP.ServiceAccountKeyPath = "" // this isn't written to disk
						Expect(appConfig.State).To(Equal(existingState))
					})
				})

				DescribeTable("when non-matching configuration is passed in",
					func(args []string, expected string) {
						_, err := c.Bootstrap(args)

						Expect(err).To(MatchError(expected))
					},
					Entry("returns an error for non-matching IAAS", []string{"bbl", "up", "--iaas", "aws"},
						"The iaas type cannot be changed for an existing environment. The current iaas type is gcp."),
					Entry("returns an error for non-matching region", []string{"bbl", "up", "--gcp-region", "some-other-region"},
						"The region cannot be changed for an existing environment. The current region is some-region."),
					Entry("returns an error for non-matching project id", []string{"bbl", "up", "--gcp-service-account-key", `{"project_id": "some-other-project-id"}`},
						"The project ID cannot be changed for an existing environment. The current project ID is some-project-id."),
				)
			})
		})

		Context("using Azure", func() {
			Context("when a previous state does not exist", func() {
				Context("when configuration is passed in by flag", func() {
					var args []string

					BeforeEach(func() {
						args = []string{
							"bbl", "up", "--name", "some-env-id",
							"--iaas", "azure",
							"--azure-client-id", "client-id",
							"--azure-client-secret", "client-secret",
							"--azure-region", "region",
							"--azure-subscription-id", "subscription-id",
							"--azure-tenant-id", "tenant-id",
						}
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State
						Expect(state.IAAS).To(Equal("azure"))
						Expect(state.Azure.ClientID).To(Equal("client-id"))
						Expect(state.Azure.ClientSecret).To(Equal("client-secret"))
						Expect(state.Azure.Region).To(Equal("region"))
						Expect(state.Azure.SubscriptionID).To(Equal("subscription-id"))
						Expect(state.Azure.TenantID).To(Equal("tenant-id"))
					})

					It("returns the command and its flags", func() {
						appConfig, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.SubcommandFlags).To(Equal(application.StringSlice{"--name", "some-env-id"}))
					})
				})

				Context("when configuration is passed in by env vars", func() {
					var args []string

					BeforeEach(func() {
						args = []string{"bbl", "up"}

						os.Setenv("BBL_IAAS", "azure")
						os.Setenv("BBL_AZURE_CLIENT_ID", "azure-client-id")
						os.Setenv("BBL_AZURE_CLIENT_SECRET", "azure-client-secret")
						os.Setenv("BBL_AZURE_REGION", "azure-region")
						os.Setenv("BBL_AZURE_SUBSCRIPTION_ID", "azure-subscription-id")
						os.Setenv("BBL_AZURE_TENANT_ID", "azure-tenant-id")
					})

					It("returns a state containing configuration", func() {
						appConfig, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State

						Expect(state.IAAS).To(Equal("azure"))
						Expect(state.Azure.ClientID).To(Equal("azure-client-id"))
						Expect(state.Azure.ClientSecret).To(Equal("azure-client-secret"))
						Expect(state.Azure.Region).To(Equal("azure-region"))
						Expect(state.Azure.SubscriptionID).To(Equal("azure-subscription-id"))
						Expect(state.Azure.TenantID).To(Equal("azure-tenant-id"))
					})

					It("returns the command", func() {
						appConfig, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.SubcommandFlags).To(Equal(application.StringSlice{}))
					})
				})
			})

			Context("when a previous state exists", func() {
				BeforeEach(func() {
					fakeStateMigrator.MigrateCall.Returns.State = storage.State{
						IAAS: "azure",
						Azure: storage.Azure{
							ClientID:       "client-id",
							ClientSecret:   "client-secret",
							Region:         "region",
							SubscriptionID: "subscription-id",
							TenantID:       "tenant-id",
						},
						EnvID: "some-env-id",
					}
				})

				Context("when no configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						appConfig, err := c.Bootstrap([]string{
							"bbl", "up",
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.State.EnvID).To(Equal("some-env-id"))

						workingDir, err := os.Getwd()
						Expect(err).NotTo(HaveOccurred())

						Expect(fakeStateBootstrap.GetStateCall.Receives.Dir).To(Equal(workingDir))
					})
				})

				Context("when valid matching configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						appConfig, err := c.Bootstrap([]string{
							"bbl", "up",
							"--iaas", "azure",
							"--azure-client-id", "client-id",
							"--azure-client-secret", "client-secret",
							"--azure-region", "region",
							"--azure-subscription-id", "subscription-id",
							"--azure-tenant-id", "tenant-id",
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.State.EnvID).To(Equal("some-env-id"))
					})
				})

				DescribeTable("when non-matching configuration is passed in",
					func(args []string, expected string) {
						_, err := c.Bootstrap(args)

						Expect(err).To(MatchError(expected))
					},
					Entry("returns an error for non-matching IAAS", []string{"bbl", "up", "--iaas", "aws"},
						"The iaas type cannot be changed for an existing environment. The current iaas type is azure."),
				)
			})
		})
	})

	Describe("ValidateIAAS", func() {
		DescribeTable("when configuration is invalid",
			func(state storage.State, expectedErr string) {
				err := config.ValidateIAAS(state)
				Expect(err).To(MatchError(expectedErr))
			},
			Entry("when IAAS is missing",
				storage.State{},
				"--iaas [gcp, aws, azure] must be provided or BBL_IAAS must be set"),
			Entry("when IAAS is unsupported",
				storage.State{
					IAAS: "not-a-real-iaas",
				},
				"--iaas [gcp, aws, azure] must be provided or BBL_IAAS must be set"),
			Entry("when AWS access key is missing",
				storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						SecretAccessKey: "some-secret-key",
						Region:          "some-region",
					},
				},
				"AWS access key ID must be provided (--aws-access-key-id or BBL_AWS_ACCESS_KEY_ID)"),
			Entry("when AWS key is missing",
				storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						AccessKeyID: "some-access-key",
						Region:      "some-region",
					},
				},
				"AWS secret access key must be provided (--aws-secret-access-key or BBL_AWS_SECRET_ACCESS_KEY)"),
			Entry("when AWS region is missing",
				storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key",
						SecretAccessKey: "some-secret-key",
					},
				},
				"AWS region must be provided (--aws-region or BBL_AWS_REGION)"),
			Entry("when GCP service account key is missing",
				storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						Region: "some-region",
					},
				},
				"GCP service account key must be provided (--gcp-service-account-key or BBL_GCP_SERVICE_ACCOUNT_KEY)"),
			Entry("when GCP region is missing",
				storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
					},
				},
				"GCP region must be provided (--gcp-region or BBL_GCP_REGION)"),
			Entry("when Azure client id is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						ClientSecret:   "some-client-secret",
						Region:         "region",
						TenantID:       "some-tenant-id",
						SubscriptionID: "some-subscription-id",
					},
				},
				"Azure client id must be provided (--azure-client-id or BBL_AZURE_CLIENT_ID)"),
			Entry("when Azure client secret is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						ClientID:       "some-client-id",
						Region:         "region",
						SubscriptionID: "some-subscription-id",
						TenantID:       "some-tenant-id",
					},
				},
				"Azure client secret must be provided (--azure-client-secret or BBL_AZURE_CLIENT_SECRET)"),
			Entry("when Azure region is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						ClientID:       "some-client-id",
						ClientSecret:   "some-client-secret",
						SubscriptionID: "some-subscription-id",
						TenantID:       "some-tenant-id",
					},
				}, "Azure region must be provided (--azure-region or BBL_AZURE_REGION)"),
			Entry("when Azure subscription is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						ClientID:     "some-client-id",
						ClientSecret: "some-client-secret",
						Region:       "region",
						TenantID:     "some-tenant-id",
					},
				},
				"Azure subscription id must be provided (--azure-subscription-id or BBL_AZURE_SUBSCRIPTION_ID)"),
			Entry("when Azure tenant id is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						ClientID:       "some-client-id",
						ClientSecret:   "some-client-secret",
						Region:         "region",
						SubscriptionID: "some-subscription-id",
					},
				},
				"Azure tenant id must be provided (--azure-tenant-id or BBL_AZURE_TENANT_ID)"),
			Entry("when vSphere vcenter user is missing",
				storage.State{
					IAAS: "vsphere",
					VSphere: storage.VSphere{
						VCenterPassword: "password",
						VCenterIP:       "ip",
						VCenterDC:       "dc",
						Cluster:         "cluster",
						VCenterRP:       "rp",
						Network:         "network",
						VCenterDS:       "ds",
						Subnet:          "subnet",
					},
				},
				"vSphere vcenter user must be provided (--vsphere-vcenter-user or BBL_VSPHERE_VCENTER_USER)"),
			Entry("when vSphere vcenter password is missing",
				storage.State{
					IAAS: "vsphere",
					VSphere: storage.VSphere{
						VCenterUser: "user",
						VCenterIP:   "ip",
						VCenterDC:   "dc",
						Cluster:     "cluster",
						VCenterRP:   "rp",
						Network:     "network",
						VCenterDS:   "ds",
						Subnet:      "subnet",
					},
				},
				"vSphere vcenter password must be provided (--vsphere-vcenter-password or BBL_VSPHERE_VCENTER_PASSWORD)"),
			Entry("when vSphere vcenter ip is missing",
				storage.State{
					IAAS: "vsphere",
					VSphere: storage.VSphere{
						VCenterUser:     "user",
						VCenterPassword: "password",
						VCenterDC:       "dc",
						Cluster:         "cluster",
						VCenterRP:       "rp",
						Network:         "network",
						VCenterDS:       "ds",
						Subnet:          "subnet",
					},
				},
				"vSphere vcenter ip must be provided (--vsphere-vcenter-ip or BBL_VSPHERE_VCENTER_IP)"),
			Entry("when vSphere vcenter datacenter is missing",
				storage.State{
					IAAS: "vsphere",
					VSphere: storage.VSphere{
						VCenterUser:     "user",
						VCenterPassword: "password",
						VCenterIP:       "ip",
						Cluster:         "cluster",
						VCenterRP:       "rp",
						Network:         "network",
						VCenterDS:       "ds",
						Subnet:          "subnet",
					},
				},
				"vSphere vcenter datacenter must be provided (--vsphere-vcenter-dc or BBL_VSPHERE_VCENTER_DC)"),
			Entry("when vSphere cluster is missing",
				storage.State{
					IAAS: "vsphere",
					VSphere: storage.VSphere{
						VCenterUser:     "user",
						VCenterPassword: "password",
						VCenterIP:       "ip",
						VCenterDC:       "dc",
						VCenterRP:       "rp",
						Network:         "network",
						VCenterDS:       "ds",
						Subnet:          "subnet",
					},
				},
				"vSphere cluster must be provided (--vsphere-vcenter-cluster or BBL_VSPHERE_VCENTER_CLUSTER)"),
			Entry("when vSphere vcenter resource pool is missing",
				storage.State{
					IAAS: "vsphere",
					VSphere: storage.VSphere{
						VCenterUser:     "user",
						VCenterPassword: "password",
						VCenterIP:       "ip",
						VCenterDC:       "dc",
						Cluster:         "cluster",
						Network:         "network",
						VCenterDS:       "ds",
						Subnet:          "subnet",
					},
				},
				"vSphere vcenter resource pool must be provided (--vsphere-vcenter-rp or BBL_VSPHERE_VCENTER_RP)"),
			Entry("when vSphere network is missing",
				storage.State{
					IAAS: "vsphere",
					VSphere: storage.VSphere{
						VCenterUser:     "user",
						VCenterPassword: "password",
						VCenterIP:       "ip",
						VCenterDC:       "dc",
						Cluster:         "cluster",
						VCenterRP:       "rp",
						VCenterDS:       "ds",
						Subnet:          "subnet",
					},
				},
				"vSphere network must be provided (--vsphere-network or BBL_VSPHERE_NETWORK)"),
			Entry("when vSphere vcenter datastore is missing",
				storage.State{
					IAAS: "vsphere",
					VSphere: storage.VSphere{
						VCenterUser:     "user",
						VCenterPassword: "password",
						VCenterIP:       "ip",
						VCenterDC:       "dc",
						Cluster:         "cluster",
						VCenterRP:       "rp",
						Network:         "network",
						Subnet:          "subnet",
					},
				},
				"vSphere vcenter datastore must be provided (--vsphere-vcenter-ds or BBL_VSPHERE_VCENTER_DS)"),
			Entry("when vSphere subnet is missing",
				storage.State{
					IAAS: "vsphere",
					VSphere: storage.VSphere{
						VCenterUser:     "user",
						VCenterPassword: "password",
						VCenterIP:       "ip",
						VCenterDC:       "dc",
						Cluster:         "cluster",
						VCenterRP:       "rp",
						Network:         "network",
						VCenterDS:       "ds",
					},
				},
				"vSphere subnet must be provided (--vsphere-subnet or BBL_VSPHERE_SUBNET)"),
		)
	})
})
