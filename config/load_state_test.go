package config_test

import (
	"errors"
	"io/ioutil"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/config"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
)

var _ = Describe("LoadState", func() {
	var c config.Config

	BeforeEach(func() {
		getState := func(string) (storage.State, error) {
			return storage.State{}, nil
		}
		c = config.NewConfig(getState)
		os.Clearenv()
	})

	Describe("Bootstrap", func() {
		Context("using AWS", func() {
			Context("when a previous state does not exist", func() {
				Context("when configuration is passed in by flag", func() {
					Context("when configuration is valid", func() {
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
							parsedFlags, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())

							state := parsedFlags.State

							Expect(state.IAAS).To(Equal("aws"))
							Expect(state.AWS.AccessKeyID).To(Equal("some-access-key"))
							Expect(state.AWS.SecretAccessKey).To(Equal("some-secret-key"))
							Expect(state.AWS.Region).To(Equal("some-region"))
						})

						It("returns the remaining arguments", func() {
							parsedFlags, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())

							Expect(parsedFlags.RemainingArgs).To(Equal([]string{"up", "--name", "some-env-id"}))
						})

						Context("when configuration includes global flags", func() {
							BeforeEach(func() {
								args = append([]string{
									"bbl",
									"--help",
									"--debug",
									"--version",
									"--state-dir", "some-state-dir",
								}, args[1:]...)
							})

							It("returns global flags", func() {
								parsedFlags, err := c.Bootstrap(args)

								Expect(err).NotTo(HaveOccurred())

								Expect(parsedFlags.Help).To(BeTrue())
								Expect(parsedFlags.Debug).To(BeTrue())
								Expect(parsedFlags.Version).To(BeTrue())
								Expect(parsedFlags.StateDir).To(Equal("some-state-dir"))
							})
						})
					})
				})

				Context("when configuration is passed in by env vars", func() {
					Context("when configuration is valid", func() {
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
							parsedFlags, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())

							state := parsedFlags.State
							Expect(state.IAAS).To(Equal("aws"))
							Expect(state.AWS.AccessKeyID).To(Equal("some-access-key-id"))
							Expect(state.AWS.SecretAccessKey).To(Equal("some-secret-key"))
							Expect(state.AWS.Region).To(Equal("some-region"))
						})

						It("returns the remaining arguments", func() {
							parsedFlags, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())

							Expect(parsedFlags.RemainingArgs).To(Equal([]string{"up"}))
						})

						Context("when configuration includes global flags", func() {
							BeforeEach(func() {
								os.Setenv("BBL_DEBUG", "true")
							})

							AfterEach(func() {
								os.Unsetenv("BBL_DEBUG")
							})

							It("returns global flags", func() {
								parsedFlags, err := c.Bootstrap(args)

								Expect(err).NotTo(HaveOccurred())

								Expect(parsedFlags.Debug).To(BeTrue())
							})
						})
					})
				})
			})

			Context("when a previous state exists", func() {
				var getStateArg string

				BeforeEach(func() {
					getState := func(dir string) (storage.State, error) {
						getStateArg = dir

						return storage.State{
							IAAS: "aws",
							AWS: storage.AWS{
								AccessKeyID:     "some-access-key-id",
								SecretAccessKey: "some-secret-access-key",
								Region:          "some-region",
							},
							EnvID: "some-env-id",
						}, nil
					}
					c = config.NewConfig(getState)
				})

				Context("when no configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						parsedFlags, err := c.Bootstrap([]string{
							"bbl",
							"create-lbs",
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(parsedFlags.State.EnvID).To(Equal("some-env-id"))

						workingDir, err := os.Getwd()
						Expect(err).NotTo(HaveOccurred())

						Expect(getStateArg).To(Equal(workingDir))
					})

					Context("when state dir is specified", func() {
						It("uses that state dir", func() {
							_, err := c.Bootstrap([]string{
								"bbl",
								"create-lbs",
								"--state-dir", "some-state-dir",
							})
							Expect(err).NotTo(HaveOccurred())

							Expect(getStateArg).To(Equal("some-state-dir"))
						})
					})
				})

				Context("when valid matching configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						parsedFlags, err := c.Bootstrap([]string{
							"bbl",
							"create-lbs",
							"--iaas", "aws",
							"--aws-access-key-id", "some-access-key-id",
							"--aws-secret-access-key", "some-secret-access-key",
							"--aws-region", "some-region",
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(parsedFlags.State.EnvID).To(Equal("some-env-id"))
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

				Context("when invalid state dir is passed in", func() {
					BeforeEach(func() {
						getState := func(string) (storage.State, error) {
							return storage.State{}, errors.New("some state dir error")
						}
						c = config.NewConfig(getState)
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

				Context("when parser errors", func() {
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
				serviceAccountKey = `{"real": "json"}`

				err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when a previous state does not exist", func() {
				Context("when configuration is passed in by flag", func() {
					Context("when configuration is valid", func() {
						var args []string

						BeforeEach(func() {
							args = []string{
								"bbl", "up", "--name", "some-env-id",
								"--iaas", "gcp",
								"--gcp-service-account-key", serviceAccountKeyPath,
								"--gcp-project-id", "some-project-id",
								"--gcp-zone", "some-availability-zone",
								"--gcp-region", "some-region",
							}
						})

						It("returns a state object containing configuration flags", func() {
							parsedFlags, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())

							state := parsedFlags.State
							Expect(state.IAAS).To(Equal("gcp"))
							Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
							Expect(state.GCP.Zone).To(Equal("some-availability-zone"))
							Expect(state.GCP.Region).To(Equal("some-region"))
						})

						It("returns the remaining arguments", func() {
							parsedFlags, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())

							Expect(parsedFlags.RemainingArgs).To(Equal([]string{"up", "--name", "some-env-id"}))
						})

						Context("when service account key is passed inline", func() {
							var args []string

							BeforeEach(func() {
								args = []string{
									"bbl", "up", "--name", "some-env-id",
									"--iaas", "gcp",
									"--gcp-service-account-key", serviceAccountKey,
									"--gcp-project-id", "some-project-id",
									"--gcp-zone", "some-availability-zone",
									"--gcp-region", "some-region",
								}
							})

							It("returns a state object containing service account key", func() {
								parsedFlags, err := c.Bootstrap(args)

								Expect(err).NotTo(HaveOccurred())

								Expect(parsedFlags.State.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
							})
						})

						Context("when configuration includes global flags", func() {
							BeforeEach(func() {
								args = append([]string{
									"bbl",
									"--help",
									"--debug",
									"--version",
									"--state-dir", "some-state-dir",
								}, args[1:]...)
							})

							It("returns global flags", func() {
								parsedFlags, err := c.Bootstrap(args)

								Expect(err).NotTo(HaveOccurred())

								Expect(parsedFlags.Help).To(BeTrue())
								Expect(parsedFlags.Debug).To(BeTrue())
								Expect(parsedFlags.Version).To(BeTrue())
								Expect(parsedFlags.StateDir).To(Equal("some-state-dir"))
							})
						})
					})

					Context("when configuration is invalid", func() {
						var args []string

						Context("when service account key file is missing", func() {
							BeforeEach(func() {
								args = []string{
									"bbl",
									"up",
									"--iaas", "gcp",
									"--gcp-service-account-key", "/this/file/isn't/real",
									"--gcp-project-id", "some-project-id",
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
								tempFile, err := ioutil.TempFile("", "invalidGcpServiceAccountKey")
								Expect(err).NotTo(HaveOccurred())

								serviceAccountKeyPath = tempFile.Name()
								serviceAccountKey = `this isn't real json`

								err = ioutil.WriteFile(serviceAccountKeyPath, []byte(serviceAccountKey), os.ModePerm)
								Expect(err).NotTo(HaveOccurred())

								args = []string{
									"bbl", "up", "--name", "some-env-id",
									"--iaas", "gcp",
									"--gcp-service-account-key", serviceAccountKeyPath,
									"--gcp-project-id", "some-project-id",
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
				})

				Context("when configuration is passed in by env vars", func() {
					var args []string

					BeforeEach(func() {
						args = []string{"bbl", "up"}

						os.Setenv("BBL_IAAS", "gcp")
						os.Setenv("BBL_GCP_SERVICE_ACCOUNT_KEY", serviceAccountKey)
						os.Setenv("BBL_GCP_PROJECT_ID", "some-project-id")
						os.Setenv("BBL_GCP_ZONE", "some-zone")
						os.Setenv("BBL_GCP_REGION", "some-region")
					})

					It("returns a state containing configuration", func() {
						parsedFlags, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						state := parsedFlags.State

						Expect(state.IAAS).To(Equal("gcp"))
						Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
						Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
						Expect(state.GCP.Zone).To(Equal("some-zone"))
						Expect(state.GCP.Region).To(Equal("some-region"))
					})

					It("returns the remaining arguments", func() {
						parsedFlags, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						Expect(parsedFlags.RemainingArgs).To(Equal([]string{"up"}))
					})

					Context("when configuration includes global flags", func() {
						BeforeEach(func() {
							os.Setenv("BBL_DEBUG", "true")
						})

						AfterEach(func() {
							os.Unsetenv("BBL_DEBUG")
						})

						It("returns global flags", func() {
							parsedFlags, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())

							Expect(parsedFlags.Debug).To(BeTrue())
						})
					})
				})
			})

			Context("when a previous state exists", func() {
				var getStateArg string

				BeforeEach(func() {
					getState := func(dir string) (storage.State, error) {
						getStateArg = dir

						return storage.State{
							IAAS: "gcp",
							GCP: storage.GCP{
								ServiceAccountKey: serviceAccountKey,
								ProjectID:         "some-project-id",
								Zone:              "some-zone",
								Region:            "some-region",
							},
							EnvID: "some-env-id",
						}, nil
					}
					c = config.NewConfig(getState)
				})

				Context("when no configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						parsedFlags, err := c.Bootstrap([]string{
							"bbl",
							"create-lbs",
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(parsedFlags.State.EnvID).To(Equal("some-env-id"))

						workingDir, err := os.Getwd()
						Expect(err).NotTo(HaveOccurred())

						Expect(getStateArg).To(Equal(workingDir))
					})

					Context("when state dir is specified", func() {
						It("uses that state dir", func() {
							_, err := c.Bootstrap([]string{
								"bbl",
								"create-lbs",
								"--state-dir", "some-state-dir",
							})
							Expect(err).NotTo(HaveOccurred())

							Expect(getStateArg).To(Equal("some-state-dir"))
						})
					})
				})

				Context("when valid matching configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						parsedFlags, err := c.Bootstrap([]string{
							"bbl",
							"create-lbs",
							"--iaas", "gcp",
							"--gcp-service-account-key", serviceAccountKey,
							"--gcp-project-id", "some-project-id",
							"--gcp-zone", "some-zone",
							"--gcp-region", "some-region",
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(parsedFlags.State.EnvID).To(Equal("some-env-id"))
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
				)
			})
		})

		Context("using Azure", func() {
			Context("when a previous state does not exist", func() {
				Context("when configuration is passed in by flag", func() {
					Context("when configuration is valid", func() {
						var args []string

						BeforeEach(func() {
							args = []string{
								"bbl", "up", "--name", "some-env-id",
								"--iaas", "azure",
								"--azure-subscription-id", "subscription-id",
								"--azure-tenant-id", "tenant-id",
								"--azure-client-id", "client-id",
								"--azure-client-secret", "client-secret",
							}
						})

						It("returns a state object containing configuration flags", func() {
							parsedFlags, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())

							state := parsedFlags.State
							Expect(state.IAAS).To(Equal("azure"))
							Expect(state.Azure.SubscriptionID).To(Equal("subscription-id"))
							Expect(state.Azure.TenantID).To(Equal("tenant-id"))
							Expect(state.Azure.ClientID).To(Equal("client-id"))
							Expect(state.Azure.ClientSecret).To(Equal("client-secret"))
						})

						It("returns the remaining arguments", func() {
							parsedFlags, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())

							Expect(parsedFlags.RemainingArgs).To(Equal([]string{"up", "--name", "some-env-id"}))
						})

						Context("when configuration includes global flags", func() {
							BeforeEach(func() {
								args = append([]string{
									"bbl",
									"--help",
									"--debug",
									"--version",
									"--state-dir", "some-state-dir",
								}, args[1:]...)
							})

							It("returns global flags", func() {
								parsedFlags, err := c.Bootstrap(args)

								Expect(err).NotTo(HaveOccurred())

								Expect(parsedFlags.Help).To(BeTrue())
								Expect(parsedFlags.Debug).To(BeTrue())
								Expect(parsedFlags.Version).To(BeTrue())
								Expect(parsedFlags.StateDir).To(Equal("some-state-dir"))
							})
						})
					})

				})

				Context("when configuration is passed in by env vars", func() {
					var args []string

					BeforeEach(func() {
						args = []string{"bbl", "up"}

						os.Setenv("BBL_IAAS", "azure")
						os.Setenv("BBL_AZURE_SUBSCRIPTION_ID", "azure-subscription-id")
						os.Setenv("BBL_AZURE_TENANT_ID", "azure-tenant-id")
						os.Setenv("BBL_AZURE_CLIENT_ID", "azure-client-id")
						os.Setenv("BBL_AZURE_CLIENT_SECRET", "azure-client-secret")
					})

					It("returns a state containing configuration", func() {
						parsedFlags, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						state := parsedFlags.State

						Expect(state.IAAS).To(Equal("azure"))
						Expect(state.Azure.SubscriptionID).To(Equal("azure-subscription-id"))
						Expect(state.Azure.TenantID).To(Equal("azure-tenant-id"))
						Expect(state.Azure.ClientID).To(Equal("azure-client-id"))
						Expect(state.Azure.ClientSecret).To(Equal("azure-client-secret"))
					})

					It("returns the remaining arguments", func() {
						parsedFlags, err := c.Bootstrap(args)

						Expect(err).NotTo(HaveOccurred())

						Expect(parsedFlags.RemainingArgs).To(Equal([]string{"up"}))
					})

					Context("when configuration includes global flags", func() {
						BeforeEach(func() {
							os.Setenv("BBL_DEBUG", "true")
						})

						AfterEach(func() {
							os.Unsetenv("BBL_DEBUG")
						})

						It("returns global flags", func() {
							parsedFlags, err := c.Bootstrap(args)

							Expect(err).NotTo(HaveOccurred())

							Expect(parsedFlags.Debug).To(BeTrue())
						})
					})
				})
			})

			Context("when a previous state exists", func() {
				var getStateArg string

				BeforeEach(func() {
					getState := func(dir string) (storage.State, error) {
						getStateArg = dir

						return storage.State{
							IAAS: "azure",
							Azure: storage.Azure{
								SubscriptionID: "subscription-id",
								TenantID:       "tenant-id",
								ClientID:       "client-id",
								ClientSecret:   "client-secret",
							},
							EnvID: "some-env-id",
						}, nil
					}
					c = config.NewConfig(getState)
				})

				Context("when no configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						parsedFlags, err := c.Bootstrap([]string{
							"bbl",
							"create-lbs",
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(parsedFlags.State.EnvID).To(Equal("some-env-id"))

						workingDir, err := os.Getwd()
						Expect(err).NotTo(HaveOccurred())

						Expect(getStateArg).To(Equal(workingDir))
					})

					Context("when state dir is specified", func() {
						It("uses that state dir", func() {
							_, err := c.Bootstrap([]string{
								"bbl",
								"create-lbs",
								"--state-dir", "some-state-dir",
							})
							Expect(err).NotTo(HaveOccurred())

							Expect(getStateArg).To(Equal("some-state-dir"))
						})
					})
				})

				Context("when valid matching configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						parsedFlags, err := c.Bootstrap([]string{
							"bbl",
							"create-lbs",
							"--iaas", "azure",
							"--azure-subscription-id", "subscription-id",
							"--azure-tenant-id", "tenant-id",
							"--azure-client-id", "client-id",
							"--azure-client-secret", "client-secret",
						})
						Expect(err).NotTo(HaveOccurred())

						Expect(parsedFlags.State.EnvID).To(Equal("some-env-id"))
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

		DescribeTable("when IAAS is not set",
			func(args []string, expectError bool, expected string) {
				_, err := c.Bootstrap(args)

				if expectError {
					Expect(err).To(MatchError(expected))
				} else {
					Expect(err).NotTo(HaveOccurred())
				}
			},
			Entry("when help flag is set", []string{"bbl", "up", "--help"}, false, ""),
			Entry("when help command is used", []string{"bbl", "help"}, false, ""),
			Entry("when no command is used", []string{"bbl"}, false, ""),
			Entry("when version flag is set", []string{"bbl", "--version"}, false, ""),
			Entry("when version command is used", []string{"bbl", "version"}, false, ""),
			// Entry("when invalid flag is passed", []string{"bbl", "--foo", "bar"}, true, "flag provided but not defined: -foo"),
		)
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
			Entry("when GCP project id is missing",
				storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: "some-service-account-key",
						Region:            "some-region",
						Zone:              "some-availability-zone",
					},
				},
				"up",
				"GCP project ID must be provided"),
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
			Entry("when Azure subscription is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						TenantID:     "some-tenant-id",
						ClientID:     "some-client-id",
						ClientSecret: "some-client-secret",
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
						SubscriptionID: "some-subscription-id",
					},
				},
				"up",
				"Azure tenant id must be provided"),
			Entry("when Azure client id is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						TenantID:       "some-tenant-id",
						ClientSecret:   "some-client-secret",
						SubscriptionID: "some-subscription-id",
					},
				},
				"up",
				"Azure client id must be provided"),
			Entry("when Azure client secret is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						TenantID:       "some-tenant-id",
						ClientID:       "some-client-id",
						SubscriptionID: "some-subscription-id",
					},
				},
				"up",
				"Azure client secret must be provided"),
		)
	})
})
