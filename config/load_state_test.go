package config_test

import (
	"errors"
	"io/ioutil"
	"os"

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
		c                  config.Config
	)

	BeforeEach(func() {
		fakeLogger = &fakes.Logger{}
		fakeStateBootstrap = &fakes.StateBootstrap{}
		c = config.NewConfig(fakeStateBootstrap, fakeLogger)
		os.Clearenv()
	})

	Describe("Bootstrap", func() {
		Describe("help and version", func() {
			Context("when no commands are passed", func() {
				It("sets the help command", func() {
					appConfig, err := c.Bootstrap([]string{"bbl"})
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("help"))
				})
			})

			Context("when help is passed as a flag", func() {
				It("sets the help command", func() {
					appConfig, err := c.Bootstrap([]string{"bbl", "--help"})
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("help"))
				})
			})

			Context("when version is passed as a flag", func() {
				It("sets the version command", func() {
					appConfig, err := c.Bootstrap([]string{"bbl", "--version"})
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("version"))
				})
			})

			DescribeTable("subcommand help for help and version",
				func(args []string, expectedCommand string) {
					appConfig, err := c.Bootstrap(args)
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal(expectedCommand))
					Expect(appConfig.ShowCommandHelp).To(BeTrue())
				},
				Entry("bbl help help", []string{"bbl", "help", "help"}, "help"),
				Entry("bbl help --help", []string{"bbl", "help", "--help"}, "help"),
				Entry("bbl help version", []string{"bbl", "help", "version"}, "version"),
				Entry("bbl version --help", []string{"bbl", "version", "--help"}, "version"),
			)
		})

		Describe("global flags", func() {
			It("returns global flags", func() {
				args := []string{
					"bbl",
					"up",
					"--debug",
					"--state-dir", "some-state-dir",
				}

				appConfig, err := c.Bootstrap(args)
				Expect(err).NotTo(HaveOccurred())

				Expect(appConfig.Command).To(Equal("up"))
				Expect(appConfig.Global.Debug).To(BeTrue())
				Expect(appConfig.Global.StateDir).To(Equal("some-state-dir"))
			})

			Context("when --help is passed in after a command", func() {
				It("returns command help", func() {
					args := []string{
						"bbl",
						"up",
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
					args := []string{
						"bbl",
						"help",
						"up",
					}

					appConfig, err := c.Bootstrap(args)
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("up"))
					Expect(appConfig.ShowCommandHelp).To(BeTrue())
				})
			})

			Context("when debug flag is passed in through environment variables", func() {
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
		})

		Describe("reading a previous state file", func() {
			BeforeEach(func() {
				fakeStateBootstrap.GetStateCall.Returns.State = storage.State{
					IAAS:  "aws",
					EnvID: "some-env-id",
				}
			})

			It("returns the existing state", func() {
				appConfig, err := c.Bootstrap([]string{
					"bbl",
					"create-lbs",
				})
				Expect(err).NotTo(HaveOccurred())

				Expect(appConfig.State.EnvID).To(Equal("some-env-id"))
			})

			It("uses the working directory", func() {
				appConfig, err := c.Bootstrap([]string{
					"bbl",
					"create-lbs",
				})
				Expect(err).NotTo(HaveOccurred())

				workingDir, err := os.Getwd()
				Expect(err).NotTo(HaveOccurred())

				Expect(fakeStateBootstrap.GetStateCall.Receives.Dir).To(Equal(workingDir))
				Expect(appConfig.Global.StateDir).To(Equal(workingDir))
			})

			Context("when state dir is specified", func() {
				It("returns state from that dir", func() {
					appConfig, err := c.Bootstrap([]string{
						"bbl",
						"create-lbs",
						"--state-dir", "some-state-dir",
					})
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeStateBootstrap.GetStateCall.Receives.Dir).To(Equal("some-state-dir"))
					Expect(appConfig.Global.StateDir).To(Equal("some-state-dir"))
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
						"create-lbs",
						"--state-dir", "/this/will/not/work",
					})

					Expect(err).To(MatchError("some state dir error"))
				})
			})

			Context("when state-dir flag is passed without an argument", func() {
				It("returns an error", func() {
					_, err := c.Bootstrap([]string{
						"bbl",
						"create-lbs",
						"--state-dir",
						"--help",
					})

					Expect(err).To(MatchError("expected argument for flag `-s, --state-dir', but got option `--help'"))
				})
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
							"bbl",
							"up",
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
					fakeStateBootstrap.GetStateCall.Returns.State = storage.State{
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
							"bbl",
							"create-lbs",
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
					Entry("returns an error for non-matching IAAS", []string{"bbl", "create-lbs", "--iaas", "gcp"},
						"The iaas type cannot be changed for an existing environment. The current iaas type is aws."),
					Entry("returns an error for non-matching region", []string{"bbl", "create-lbs", "--aws-region", "some-other-region"},
						"The region cannot be changed for an existing environment. The current region is some-region."),
				)
			})
		})

		Context("using GCP", func() {
			var (
				serviceAccountKeyPath string
				serviceAccountKey     string
			)
			BeforeEach(func() {
				tempFile, err := ioutil.TempFile("", "gcpServiceAccountKey")
				Expect(err).NotTo(HaveOccurred())

				serviceAccountKeyPath = tempFile.Name()
				serviceAccountKey = `{"project_id": "some-project-id"}`

				err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when a previous state does not exist", func() {
				Context("when configuration is passed in by flag", func() {
					var args []string

					BeforeEach(func() {
						args = []string{
							"bbl", "up", "--name", "some-env-id",
							"--iaas", "gcp",
							"--gcp-service-account-key", serviceAccountKeyPath,
							"--gcp-zone", "some-availability-zone",
							"--gcp-region", "some-region",
						}
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State
						Expect(state.IAAS).To(Equal("gcp"))
						Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
						Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
						Expect(state.GCP.Zone).To(Equal("some-availability-zone"))
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
							args = []string{
								"bbl", "up", "--name", "some-env-id",
								"--iaas", "gcp",
								"--gcp-service-account-key", serviceAccountKey,
								"--gcp-zone", "some-availability-zone",
								"--gcp-region", "some-region",
							}
						})

						It("returns a state object containing service account key", func() {
							appConfig, err := c.Bootstrap(args)
							Expect(err).NotTo(HaveOccurred())

							Expect(appConfig.State.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
							Expect(appConfig.State.GCP.ProjectID).To(Equal("some-project-id"))
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
									"--gcp-zone", "some-availability-zone",
									"--gcp-region", "some-region",
								}
							})

							It("returns an error", func() {
								_, err := c.Bootstrap(args)

								Expect(err).To(MatchError("error unmarshalling service account key (must be valid json): invalid character '/' looking for beginning of value"))
							})
						})

						Context("when service account key is invalid json", func() {
							BeforeEach(func() {
								serviceAccountKey = `this isn't real json`
								args = []string{
									"bbl", "up", "--name", "some-env-id",
									"--iaas", "gcp",
									"--gcp-service-account-key", serviceAccountKey,
									"--gcp-zone", "some-availability-zone",
									"--gcp-region", "some-region",
								}
							})

							It("returns an error", func() {
								_, err := c.Bootstrap(args)
								Expect(err).To(MatchError(ContainSubstring("error unmarshalling service account key (must be valid json):")))
							})
						})
					})

					Context("when service account key is missing project ID field", func() {
						BeforeEach(func() {
							serviceAccountKey = `{"missing": "project_id"}`
							args = []string{
								"bbl", "up", "--name", "some-env-id",
								"--iaas", "gcp",
								"--gcp-service-account-key", serviceAccountKey,
								"--gcp-zone", "some-availability-zone",
								"--gcp-region", "some-region",
							}
						})

						It("returns an error", func() {
							_, err := c.Bootstrap(args)
							Expect(err).To(MatchError("service account key is missing field `project_id`"))
						})
					})
				})

				Context("when configuration is passed in by env vars", func() {
					var args []string

					BeforeEach(func() {
						args = []string{"bbl", "up"}

						os.Setenv("BBL_IAAS", "gcp")
						os.Setenv("BBL_GCP_SERVICE_ACCOUNT_KEY", serviceAccountKey)
						os.Setenv("BBL_GCP_ZONE", "some-zone")
						os.Setenv("BBL_GCP_REGION", "some-region")
					})

					It("returns a state containing configuration", func() {
						appConfig, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State

						Expect(state.IAAS).To(Equal("gcp"))
						Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
						Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
						Expect(state.GCP.Zone).To(Equal("some-zone"))
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
				BeforeEach(func() {
					fakeStateBootstrap.GetStateCall.Returns.State = storage.State{
						IAAS: "gcp",
						GCP: storage.GCP{
							ServiceAccountKey: serviceAccountKey,
							ProjectID:         "some-project-id",
							Zone:              "some-zone",
							Region:            "some-region",
						},
						EnvID: "some-env-id",
					}
				})

				Context("when valid matching configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						appConfig, err := c.Bootstrap([]string{
							"bbl",
							"create-lbs",
							"--iaas", "gcp",
							"--gcp-service-account-key", serviceAccountKey,
							"--gcp-zone", "some-zone",
							"--gcp-region", "some-region",
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
					Entry("returns an error for non-matching IAAS", []string{"bbl", "create-lbs", "--iaas", "aws"},
						"The iaas type cannot be changed for an existing environment. The current iaas type is gcp."),
					Entry("returns an error for non-matching region", []string{"bbl", "create-lbs", "--gcp-region", "some-other-region"},
						"The region cannot be changed for an existing environment. The current region is some-region."),
					Entry("returns an error for non-matching zone", []string{"bbl", "create-lbs", "--gcp-zone", "some-other-zone"},
						"The zone cannot be changed for an existing environment. The current zone is some-zone."),
					Entry("returns an error for non-matching project id", []string{"bbl", "create-lbs", "--gcp-service-account-key", `{"project_id": "some-other-project-id"}`},
						"The project ID cannot be changed for an existing environment. The current project ID is some-project-id."),
				)
			})

			Describe("deprecated flags", func() {
				var args []string
				Context("when the deprecated --gcp-project-id is passed in", func() {
					BeforeEach(func() {
						args = []string{
							"bbl", "up",
							"--iaas", "gcp",
							"--gcp-project-id", "ignored-project-id",
							"--gcp-service-account-key", serviceAccountKey,
						}
					})
					It("ignores the flag and prints a warning", func() {
						appConfig, err := c.Bootstrap(args)
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.State.GCP.ProjectID).To(Equal("some-project-id"))
						Expect(fakeLogger.PrintlnCall.Receives.Message).To(Equal("Deprecation warning: the --gcp-project-id (BBL_GCP_PROJECT_ID) flag is now ignored."))
					})
				})
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
							"--azure-location", "location",
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
						Expect(state.Azure.Location).To(Equal("location"))
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
						os.Setenv("BBL_AZURE_LOCATION", "azure-location")
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
						Expect(state.Azure.Location).To(Equal("azure-location"))
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
					fakeStateBootstrap.GetStateCall.Returns.State = storage.State{
						IAAS: "azure",
						Azure: storage.Azure{
							ClientID:       "client-id",
							ClientSecret:   "client-secret",
							Location:       "location",
							SubscriptionID: "subscription-id",
							TenantID:       "tenant-id",
						},
						EnvID: "some-env-id",
					}
				})

				Context("when no configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						appConfig, err := c.Bootstrap([]string{
							"bbl",
							"create-lbs",
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
							"bbl",
							"create-lbs",
							"--iaas", "azure",
							"--azure-client-id", "client-id",
							"--azure-client-secret", "client-secret",
							"--azure-location", "location",
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
					Entry("returns an error for non-matching IAAS", []string{"bbl", "create-lbs", "--iaas", "aws"},
						"The iaas type cannot be changed for an existing environment. The current iaas type is azure."),
				)
			})
		})
	})

	Describe("ValidateIAAS", func() {
		DescribeTable("when configuration is invalid",
			func(state storage.State, command string, expectedErr string) {
				err := config.ValidateIAAS(state, command)

				Expect(err).To(MatchError(expectedErr))
			},

			Entry("when IAAS is missing",
				storage.State{},
				"up",
				"--iaas [gcp, aws, azure] must be provided or BBL_IAAS must be set"),
			Entry("when IAAS is unsupported",
				storage.State{
					IAAS: "not-a-real-iaas",
				},
				"up",
				"--iaas [gcp, aws, azure] must be provided or BBL_IAAS must be set"),
			Entry("when AWS access key is missing",
				storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						SecretAccessKey: "some-secret-key",
						Region:          "some-region",
					},
				},
				"up",
				"AWS access key ID must be provided"),
			Entry("when AWS key is missing",
				storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						AccessKeyID: "some-access-key",
						Region:      "some-region",
					},
				},
				"up",
				"AWS secret access key must be provided"),
			Entry("when AWS region is missing",
				storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						AccessKeyID:     "some-access-key",
						SecretAccessKey: "some-secret-key",
					},
				},
				"up",
				"AWS region must be provided"),
			Entry("when GCP service account key is missing",
				storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ProjectID: "some-project-id",
						Zone:      "some-availability-zone",
						Region:    "some-region",
					},
				},
				"up",
				"GCP service account key must be provided"),
			Entry("when GCP region is missing",
				storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-service-account-key",
						Zone:              "some-availability-zone",
					},
				},
				"up",
				"GCP region must be provided"),
			Entry("when GCP zone is missing",
				storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ProjectID:         "some-project-id",
						ServiceAccountKey: "some-service-account-key",
						Region:            "some-region",
					},
				},
				"up",
				"GCP zone must be provided"),
			Entry("when Azure client id is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						ClientSecret:   "some-client-secret",
						Location:       "location",
						TenantID:       "some-tenant-id",
						SubscriptionID: "some-subscription-id",
					},
				},
				"up",
				"Azure client id must be provided"),
			Entry("when Azure client secret is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						ClientID:       "some-client-id",
						Location:       "location",
						SubscriptionID: "some-subscription-id",
						TenantID:       "some-tenant-id",
					},
				},
				"up",
				"Azure client secret must be provided"),
			Entry("when Azure location is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						ClientID:       "some-client-id",
						ClientSecret:   "some-client-secret",
						SubscriptionID: "some-subscription-id",
						TenantID:       "some-tenant-id",
					},
				},
				"up",
				"Azure location must be provided"),
			Entry("when Azure subscription is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						ClientID:     "some-client-id",
						ClientSecret: "some-client-secret",
						Location:     "location",
						TenantID:     "some-tenant-id",
					},
				},
				"up",
				"Azure subscription id must be provided"),
			Entry("when Azure tenant id is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						ClientID:       "some-client-id",
						ClientSecret:   "some-client-secret",
						Location:       "location",
						SubscriptionID: "some-subscription-id",
					},
				},
				"up",
				"Azure tenant id must be provided"),
		)
	})
})
