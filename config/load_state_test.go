package config_test

import (
	"errors"
	"fmt"
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

func bootstrapArgs(args []string) (config.GlobalFlags, []string, int) {
	globals, remaining, err := config.ParseArgs(args)
	Expect(err).NotTo(HaveOccurred())

	return globals, remaining, len(args)
}

var _ = Describe("LoadState", func() {
	var (
		fakeLogger         *fakes.Logger
		fakeStateBootstrap *fakes.StateBootstrap
		fakeStateMigrator  *fakes.StateMigrator
		fakeFileIO         *fakes.FileIO
		fakeDownloader     *fakes.Downloader
		c                  config.Config
	)

	BeforeEach(func() {
		fakeLogger = &fakes.Logger{}
		fakeStateBootstrap = &fakes.StateBootstrap{}
		fakeStateMigrator = &fakes.StateMigrator{}
		fakeFileIO = &fakes.FileIO{}
		fakeDownloader = &fakes.Downloader{}
		os.Clearenv()

		c = config.NewConfig(fakeStateBootstrap, fakeStateMigrator, config.NewMerger(fakeFileIO), fakeDownloader, fakeLogger, fakeFileIO)
	})

	AfterEach(func() {
		os.Clearenv()
	})

	Describe("Bootstrap", func() {
		Describe("help and version", func() {
			Context("when no commands are passed", func() {
				It("sets the help command", func() {
					appConfig, err := c.Bootstrap(bootstrapArgs([]string{"bbl"}))
					// appConfig, err := c.Bootstrap(bootstrapArgs([]string{"bbl"}))
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("help"))
					Expect(fakeStateBootstrap.GetStateCall.CallCount).To(Equal(0))
					Expect(fakeStateMigrator.MigrateCall.CallCount).To(Equal(0))
				})
			})

			Context("when --help is passed as a flag", func() {
				It("sets the help command", func() {
					appConfig, err := c.Bootstrap(bootstrapArgs([]string{"bbl", "--help"}))
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("help"))
					Expect(appConfig.ShowCommandHelp).To(BeFalse())
					Expect(fakeStateBootstrap.GetStateCall.CallCount).To(Equal(0))
					Expect(fakeStateMigrator.MigrateCall.CallCount).To(Equal(0))
				})
			})

			Context("when version is passed as a flag", func() {
				It("sets the version command", func() {
					appConfig, err := c.Bootstrap(bootstrapArgs([]string{"bbl", "--version"}))
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("version"))
					Expect(fakeStateBootstrap.GetStateCall.CallCount).To(Equal(0))
					Expect(fakeStateMigrator.MigrateCall.CallCount).To(Equal(0))
				})
			})

			DescribeTable("subcommand help for help and version",
				func(args []string, expectedCommand string) {
					appConfig, err := c.Bootstrap(bootstrapArgs(args))
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
					"bbl", "print-env",
					"--debug",
					"--state-dir", "some-state-dir",
				}

				appConfig, err := c.Bootstrap(bootstrapArgs(args))
				Expect(err).NotTo(HaveOccurred())

				Expect(appConfig.Command).To(Equal("print-env"))
				Expect(appConfig.Global.Debug).To(BeTrue())
				Expect(appConfig.Global.StateDir).To(Equal(fullStateDirPath))
			})

			Context("when --help is passed in after a command", func() {
				It("returns command help", func() {
					args := []string{
						"bbl", "print-env",
						"--help",
					}

					appConfig, err := c.Bootstrap(bootstrapArgs(args))
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("print-env"))
					Expect(appConfig.ShowCommandHelp).To(BeTrue())
				})
			})

			Context("when help is passed in before a command", func() {
				It("returns command help", func() {
					args := []string{"bbl", "help", "print-env"}

					appConfig, err := c.Bootstrap(bootstrapArgs(args))
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Command).To(Equal("print-env"))
					Expect(appConfig.ShowCommandHelp).To(BeTrue())
				})
			})

			Context("when debug flag is passed in through environment variable", func() {
				BeforeEach(func() {
					os.Setenv("BBL_DEBUG", "true")
				})

				It("returns global flags", func() {
					appConfig, err := c.Bootstrap(bootstrapArgs([]string{"bbl", "print-env"}))
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Global.Debug).To(BeTrue())
				})
			})

			Context("when state dir flag is passed in through environment variable", func() {
				BeforeEach(func() {
					os.Setenv("BBL_STATE_DIRECTORY", "/path/to/state")
				})

				It("returns global flags", func() {
					appConfig, err := c.Bootstrap(bootstrapArgs([]string{"bbl", "print-env"}))
					Expect(err).NotTo(HaveOccurred())

					Expect(appConfig.Global.StateDir).To(Equal("/path/to/state"))
				})
			})

			Context("when an external bbl-state is specified", func() {
				It("downloads the bbl state", func() {
					_, err := c.Bootstrap(bootstrapArgs([]string{
						"bbl", "print-env",
						"--state-bucket", "some-state-bucket",
						"--name", "some-name",
						"--aws-access-key-id", "some-aws-access-key",
						"--aws-secret-access-key", "some-aws-secret-access-key",
					}))
					Expect(err).NotTo(HaveOccurred())

					Expect(fakeDownloader.DownloadCall.CallCount).To(Equal(1))
					flags := fakeDownloader.DownloadCall.Receives.GlobalFlags
					Expect(flags.StateBucket).To(Equal("some-state-bucket"))
					Expect(flags.EnvID).To(Equal("some-name"))
					Expect(flags.AWSAccessKeyID).To(Equal("some-aws-access-key"))
					Expect(flags.AWSSecretAccessKey).To(Equal("some-aws-secret-access-key"))
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
				appConfig, err := c.Bootstrap(bootstrapArgs([]string{
					"bbl",
					"print-env",
				}))
				Expect(err).NotTo(HaveOccurred())

				By("migrating the existing state file", func() {
					Expect(fakeStateBootstrap.GetStateCall.CallCount).To(Equal(1))
					Expect(fakeStateMigrator.MigrateCall.CallCount).To(Equal(1))
					Expect(fakeStateMigrator.MigrateCall.Receives.State).To(Equal(gotState))
				})

				Expect(appConfig.State).To(Equal(migratedState))
			})

			It("uses the working directory", func() {
				appConfig, err := c.Bootstrap(bootstrapArgs([]string{
					"bbl",
					"print-env",
				}))
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
					appConfig, err := c.Bootstrap(bootstrapArgs([]string{
						"bbl",
						"print-env",
						"--state-dir", "some-state-dir",
					}))
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
					appConfig, err := c.Bootstrap(bootstrapArgs([]string{
						"bbl",
						"print-env",
						"--state-dir", stateDir,
					}))
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
				})

				It("returns an error", func() {
					_, err := c.Bootstrap(bootstrapArgs([]string{"bbl", "print-env", "--state-dir", "/this/will/not/work"}))

					Expect(err).To(MatchError("some state dir error"))
				})
			})

			Context("when migrating the state fails", func() {
				BeforeEach(func() {
					fakeStateMigrator.MigrateCall.Returns.Error = errors.New("coconut")
				})
				It("returns an error", func() {
					_, err := c.Bootstrap(bootstrapArgs([]string{"bbl", "print-env", "--state-dir", "some-state-dir"}))
					Expect(err).To(MatchError("coconut"))
				})
			})

			Context("when state-dir flag is passed without an argument", func() {
				It("returns an error", func() {
					_, _, err := config.ParseArgs([]string{"bbl", "print-env", "--state-dir", "--help"})

					Expect(err).To(MatchError("expected argument for flag `-s, --state-dir', but got option `--help'"))
				})
			})
		})

		Context("using Openstack", func() {
			Context("when a previous state does not exist", func() {
				Context("when configuration is passed in by flag", func() {
					var args []string

					BeforeEach(func() {
						args = []string{
							"bbl",
							"--iaas", "openstack",
							"--openstack-auth-url", "auth-url",
							"--openstack-az", "az",
							"--openstack-network-id", "network-id",
							"--openstack-network-name", "network-name",
							"--openstack-password", "password",
							"--openstack-username", "username",
							"--openstack-project", "project",
							"--openstack-domain", "domain",
							"--openstack-region", "region",
							"--openstack-cacert-file", "/path/to/file",
							"--openstack-insecure", "true",
							"--openstack-dns-name-server", "8.8.8.8",
							"--openstack-dns-name-server", "9.9.9.9",
							"up",
							"--name", "some-env-id",
						}
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State

						Expect(state.IAAS).To(Equal("openstack"))
						Expect(state.OpenStack.AuthURL).To(Equal("auth-url"))
						Expect(state.OpenStack.AZ).To(Equal("az"))
						Expect(state.OpenStack.NetworkID).To(Equal("network-id"))
						Expect(state.OpenStack.NetworkName).To(Equal("network-name"))
						Expect(state.OpenStack.Password).To(Equal("password"))
						Expect(state.OpenStack.Username).To(Equal("username"))
						Expect(state.OpenStack.Project).To(Equal("project"))
						Expect(state.OpenStack.Domain).To(Equal("domain"))
						Expect(state.OpenStack.Region).To(Equal("region"))
						Expect(state.OpenStack.CACertFile).To(Equal("/path/to/file"))
						Expect(state.OpenStack.Insecure).To(Equal("true"))
						Expect(state.OpenStack.DNSNameServers).To(Equal([]string{"8.8.8.8", "9.9.9.9"}))
					})

					It("returns the remaining arguments", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.Global.Name).To(Equal("some-env-id"))
					})
				})

				Context("when configuration is passed in by env vars", func() {
					var args []string

					BeforeEach(func() {
						args = []string{"bbl", "up"}

						os.Setenv("BBL_IAAS", "openstack")
						os.Setenv("BBL_OPENSTACK_AUTH_URL", "auth-url")
						os.Setenv("BBL_OPENSTACK_AZ", "az")
						os.Setenv("BBL_OPENSTACK_NETWORK_ID", "network-id")
						os.Setenv("BBL_OPENSTACK_NETWORK_NAME", "network-name")
						os.Setenv("BBL_OPENSTACK_PASSWORD", "password")
						os.Setenv("BBL_OPENSTACK_USERNAME", "username")
						os.Setenv("BBL_OPENSTACK_PROJECT", "project")
						os.Setenv("BBL_OPENSTACK_DOMAIN", "domain")
						os.Setenv("BBL_OPENSTACK_REGION", "region")
						os.Setenv("BBL_OPENSTACK_CACERT_FILE", "/path/to/file")
						os.Setenv("BBL_OPENSTACK_INSECURE", "true")
						os.Setenv("BBL_OPENSTACK_DNS_NAME_SERVERS", "8.8.8.8,9.9.9.9")
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State
						Expect(state.IAAS).To(Equal("openstack"))
						Expect(state.OpenStack.AuthURL).To(Equal("auth-url"))
						Expect(state.OpenStack.AZ).To(Equal("az"))
						Expect(state.OpenStack.NetworkID).To(Equal("network-id"))
						Expect(state.OpenStack.NetworkName).To(Equal("network-name"))
						Expect(state.OpenStack.Password).To(Equal("password"))
						Expect(state.OpenStack.Username).To(Equal("username"))
						Expect(state.OpenStack.Project).To(Equal("project"))
						Expect(state.OpenStack.Domain).To(Equal("domain"))
						Expect(state.OpenStack.Region).To(Equal("region"))
						Expect(state.OpenStack.CACertFile).To(Equal("/path/to/file"))
						Expect(state.OpenStack.Insecure).To(Equal("true"))
						Expect(state.OpenStack.DNSNameServers).To(Equal([]string{"8.8.8.8", "9.9.9.9"}))
					})

					It("returns the command", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
					})
				})
			})

			Context("when a previous state exists", func() {
				BeforeEach(func() {
					fakeStateMigrator.MigrateCall.Returns.State = storage.State{
						IAAS: "openstack",
						OpenStack: storage.OpenStack{
							AuthURL:        "auth-url",
							AZ:             "az",
							NetworkID:      "network-id",
							NetworkName:    "network-name",
							Password:       "password",
							Username:       "username",
							Project:        "project",
							Domain:         "domain",
							Region:         "region",
							DNSNameServers: []string{"8.8.8.8", "9.9.9.9"},
						},
						EnvID: "some-env-id",
					}
				})

				Context("when valid matching configuration is passed in", func() {
					var (
						appConfig application.Configuration
						err       error
					)
					BeforeEach(func() {
						appConfig, err = c.Bootstrap(bootstrapArgs([]string{
							"bbl", "up",
							"--iaas", "openstack",
							"--openstack-auth-url", "auth-url",
							"--openstack-az", "az",
							"--openstack-network-id", "network-id",
							"--openstack-network-name", "network-name",
							"--openstack-password", "password",
							"--openstack-username", "username",
							"--openstack-project", "project",
							"--openstack-domain", "domain",
							"--openstack-region", "new-region",
							"--openstack-dns-name-server", "1.1.1.1",
						}))
					})

					It("returns state with existing configuration", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(appConfig.State.EnvID).To(Equal("some-env-id"))
					})

					It("overrides existing configuration", func() {
						Expect(err).NotTo(HaveOccurred())
						Expect(appConfig.State.OpenStack.Region).To(Equal("new-region"))
						Expect(appConfig.State.OpenStack.DNSNameServers).To(Equal([]string{"1.1.1.1"}))
					})
				})

				DescribeTable("when non-matching configuration is passed in",
					func(args []string, expected string) {
						_, err := c.Bootstrap(bootstrapArgs(args))

						Expect(err).To(MatchError(expected))
					},
					Entry("returns an error for non-matching IAAS", []string{"bbl", "up", "--iaas", "gcp"},
						"The iaas type cannot be changed for an existing environment. The current iaas type is openstack."),
				)
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
							"--vsphere-subnet-cidr", "subnet",
							"--vsphere-vcenter-disks", "disks",
							"--vsphere-vcenter-templates", "templates",
							"--vsphere-vcenter-vms", "vms",
							"up",
							"--name", "some-env-id",
						}
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State

						Expect(state.IAAS).To(Equal("vsphere"))
						Expect(state.VSphere.VCenterUser).To(Equal("user"))
						Expect(state.VSphere.VCenterPassword).To(Equal("password"))
						Expect(state.VSphere.VCenterIP).To(Equal("ip"))
						Expect(state.VSphere.VCenterDC).To(Equal("dc"))
						Expect(state.VSphere.VCenterCluster).To(Equal("cluster"))
						Expect(state.VSphere.VCenterRP).To(Equal("rp"))
						Expect(state.VSphere.VCenterDS).To(Equal("ds"))
						Expect(state.VSphere.Network).To(Equal("network"))
						Expect(state.VSphere.SubnetCIDR).To(Equal("subnet"))
						Expect(state.VSphere.VCenterDisks).To(Equal("disks"))
						Expect(state.VSphere.VCenterTemplates).To(Equal("templates"))
						Expect(state.VSphere.VCenterVMs).To(Equal("vms"))

					})

					It("returns the remaining arguments", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.Global.Name).To(Equal("some-env-id"))
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
						os.Setenv("BBL_VSPHERE_SUBNET_CIDR", "subnet")
						os.Setenv("BBL_VSPHERE_VCENTER_DISKS", "disks")
						os.Setenv("BBL_VSPHERE_VCENTER_TEMPLATES", "templates")
						os.Setenv("BBL_VSPHERE_VCENTER_VMS", "vms")
					})

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State
						Expect(state.IAAS).To(Equal("vsphere"))
						Expect(state.VSphere.VCenterUser).To(Equal("user"))
						Expect(state.VSphere.VCenterPassword).To(Equal("password"))
						Expect(state.VSphere.VCenterIP).To(Equal("ip"))
						Expect(state.VSphere.VCenterDC).To(Equal("dc"))
						Expect(state.VSphere.VCenterCluster).To(Equal("cluster"))
						Expect(state.VSphere.VCenterRP).To(Equal("rp"))
						Expect(state.VSphere.Network).To(Equal("network"))
						Expect(state.VSphere.VCenterDS).To(Equal("ds"))
						Expect(state.VSphere.SubnetCIDR).To(Equal("subnet"))
						Expect(state.VSphere.VCenterDisks).To(Equal("disks"))
						Expect(state.VSphere.VCenterTemplates).To(Equal("templates"))
						Expect(state.VSphere.VCenterVMs).To(Equal("vms"))
					})

					It("returns the command", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
					})
				})

				Context("when configuration is passsed in by flags missing templates, disks and vms folders", func() {
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
							"--vsphere-subnet-cidr", "subnet",
							"up",
							"--name", "some-env-id",
						}
					})
					It("generates vm, templates and disks folder names", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State
						Expect(state.VSphere.VCenterDisks).To(Equal("network"))
						Expect(state.VSphere.VCenterTemplates).To(Equal("network_templates"))
						Expect(state.VSphere.VCenterVMs).To(Equal("network_vms"))
					})
				})
			})

			Context("when a previous state exists", func() {
				BeforeEach(func() {
					fakeStateMigrator.MigrateCall.Returns.State = storage.State{
						IAAS: "vsphere",
						VSphere: storage.VSphere{
							VCenterUser:      "user",
							VCenterPassword:  "password",
							VCenterIP:        "ip",
							VCenterDC:        "dc",
							VCenterCluster:   "cluster",
							VCenterRP:        "rp",
							Network:          "network",
							VCenterDS:        "ds",
							SubnetCIDR:       "subnet",
							VCenterDisks:     "disks",
							VCenterTemplates: "templates",
							VCenterVMs:       "vms",
						},
						EnvID: "some-env-id",
					}
				})

				Context("when valid matching configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs([]string{
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
							"--vsphere-subnet-cidr", "subnet",
							"--vsphere-vcenter-vms", "vms",
							"--vsphere-vcenter-templates", "templates",
							"--vsphere-vcenter-disks", "disks",
						}))
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.State.EnvID).To(Equal("some-env-id"))
					})
				})

				DescribeTable("when non-matching configuration is passed in",
					func(args []string, expected string) {
						_, err := c.Bootstrap(bootstrapArgs(args))

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
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State

						Expect(state.IAAS).To(Equal("aws"))
						Expect(state.AWS.AccessKeyID).To(Equal("some-access-key"))
						Expect(state.AWS.SecretAccessKey).To(Equal("some-secret-key"))
						Expect(state.AWS.Region).To(Equal("some-region"))
					})

					It("returns the remaining arguments", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.Global.Name).To(Equal("some-env-id"))
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

					It("returns a state object containing configuration flags", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State
						Expect(state.IAAS).To(Equal("aws"))
						Expect(state.AWS.AccessKeyID).To(Equal("some-access-key-id"))
						Expect(state.AWS.SecretAccessKey).To(Equal("some-secret-key"))
						Expect(state.AWS.Region).To(Equal("some-region"))
					})

					It("returns the command", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
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
						appConfig, err := c.Bootstrap(bootstrapArgs([]string{
							"bbl", "up",
							"--iaas", "aws",
							"--aws-access-key-id", "some-access-key-id",
							"--aws-secret-access-key", "some-secret-access-key",
							"--aws-region", "some-region",
						}))
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.State.EnvID).To(Equal("some-env-id"))
					})
				})

				DescribeTable("when non-matching configuration is passed in",
					func(args []string, expected string) {
						_, err := c.Bootstrap(bootstrapArgs(args))

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
				serviceAccountKey string
				tempFile          *os.File
			)
			BeforeEach(func() {
				var err error
				tempFile, err = ioutil.TempFile("", "temp")
				Expect(err).NotTo(HaveOccurred())
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
						appConfig, err := c.Bootstrap(bootstrapArgs(args))

						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State
						Expect(state.IAAS).To(Equal("gcp"))
						Expect(state.GCP.ServiceAccountKeyPath).To(Equal("/path/to/service/account/key"))
						Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
						Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
						Expect(state.GCP.Region).To(Equal("some-region"))
					})

					It("returns the command and its flags", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))

						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.Global.Name).To(Equal("some-env-id"))
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
							appConfig, err := c.Bootstrap(bootstrapArgs(args))

							Expect(err).NotTo(HaveOccurred())
							Expect(appConfig.State.GCP.ProjectID).To(Equal("some-project-id"))
							Expect(fakeFileIO.WriteFileCall.Receives[0].Filename).To(Equal(tempFile.Name()))
							Expect(fakeFileIO.WriteFileCall.Receives[0].Contents).To(Equal([]byte(serviceAccountKey)))
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
								_, err := c.Bootstrap(bootstrapArgs(args))

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
								_, err := c.Bootstrap(bootstrapArgs(args))
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
							_, err := c.Bootstrap(bootstrapArgs(args))
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
							_, err := c.Bootstrap(bootstrapArgs(args))
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
							fakeFileIO.WriteFileCall.Returns = []fakes.WriteFileReturn{{Error: errors.New("coconut")}}
						})

						It("returns an error", func() {
							_, err := c.Bootstrap(bootstrapArgs(args))
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
						appConfig, err := c.Bootstrap(bootstrapArgs(args))

						Expect(err).NotTo(HaveOccurred())

						state := appConfig.State

						Expect(state.IAAS).To(Equal("gcp"))
						Expect(state.GCP.ServiceAccountKeyPath).To(Equal(tempFile.Name()))
						Expect(state.GCP.ServiceAccountKey).To(Equal(serviceAccountKey))
						Expect(state.GCP.ProjectID).To(Equal("some-project-id"))
						Expect(state.GCP.Region).To(Equal("some-region"))
					})

					It("returns the remaining arguments", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs(args))
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
						appConfig, err := c.Bootstrap(bootstrapArgs([]string{
							"bbl", "up",
							"--iaas", "gcp",
							"--gcp-service-account-key", serviceAccountKey,
							"--gcp-region", "some-region",
						}))
						Expect(err).NotTo(HaveOccurred())

						appConfig.State.GCP.ServiceAccountKey = ""     // this isn't written to disk
						appConfig.State.GCP.ServiceAccountKeyPath = "" // this isn't written to disk
						Expect(appConfig.State).To(Equal(existingState))
					})
				})

				DescribeTable("when non-matching configuration is passed in",
					func(args []string, expected string) {
						_, err := c.Bootstrap(bootstrapArgs(args))

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
						appConfig, err := c.Bootstrap(bootstrapArgs(args))

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
						appConfig, err := c.Bootstrap(bootstrapArgs(args))

						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.Global.Name).To(Equal("some-env-id"))
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
						appConfig, err := c.Bootstrap(bootstrapArgs(args))

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
						appConfig, err := c.Bootstrap(bootstrapArgs(args))

						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.Command).To(Equal("up"))
						Expect(appConfig.SubcommandFlags).To(Equal(application.StringSlice{}))
					})
				})
			})

			Context("using CloudStack", func() {
				Context("when a previous state does not exist", func() {
					Context("when configuration is passed in by flag", func() {
						var args []string

						BeforeEach(func() {
							args = []string{
								"bbl",
								"--iaas", "cloudstack",
								"--cloudstack-endpoint", "http://my-cloudstack.com/client/api",
								"--cloudstack-api-key", "some-api-key",
								"--cloudstack-secret-access-key", "some-secret-key",
								"--cloudstack-zone", "some-zone",
								"--cloudstack-secure",
								"--cloudstack-iso-segment",
								"up",
								"--name", "some-env-id",
							}
						})

						It("returns a state object containing configuration flags", func() {
							appConfig, err := c.Bootstrap(bootstrapArgs(args))
							Expect(err).NotTo(HaveOccurred())

							state := appConfig.State

							Expect(state.IAAS).To(Equal("cloudstack"))
							Expect(state.CloudStack.ApiKey).To(Equal("some-api-key"))
							Expect(state.CloudStack.SecretAccessKey).To(Equal("some-secret-key"))
							Expect(state.CloudStack.Zone).To(Equal("some-zone"))
							Expect(state.CloudStack.Endpoint).To(Equal("http://my-cloudstack.com/client/api"))
							Expect(state.CloudStack.Secure).To(BeTrue())
							Expect(state.CloudStack.IsoSegment).To(BeTrue())
						})

						It("returns the remaining arguments", func() {
							appConfig, err := c.Bootstrap(bootstrapArgs(args))
							Expect(err).NotTo(HaveOccurred())

							Expect(appConfig.Command).To(Equal("up"))
							Expect(appConfig.Global.Name).To(Equal("some-env-id"))
						})
					})

					Context("when configuration is passed in by env vars", func() {
						var args []string

						BeforeEach(func() {
							args = []string{
								"bbl", "up",
							}

							os.Setenv("BBL_IAAS", "cloudstack")
							os.Setenv("BBL_CLOUDSTACK_API_KEY", "some-api-key")
							os.Setenv("BBL_CLOUDSTACK_SECRET_ACCESS_KEY", "some-secret-key")
							os.Setenv("BBL_CLOUDSTACK_ZONE", "some-zone")
							os.Setenv("BBL_CLOUDSTACK_SECURE", "true")
							os.Setenv("BBL_CLOUDSTACK_ISO_SEGMENT", "true")
							os.Setenv("BBL_CLOUDSTACK_ENDPOINT", "http://my-cloudstack.com/client/api")
						})

						It("returns a state object containing configuration flags", func() {
							appConfig, err := c.Bootstrap(bootstrapArgs(args))
							Expect(err).NotTo(HaveOccurred())

							state := appConfig.State
							Expect(state.IAAS).To(Equal("cloudstack"))
							Expect(state.CloudStack.ApiKey).To(Equal("some-api-key"))
							Expect(state.CloudStack.SecretAccessKey).To(Equal("some-secret-key"))
							Expect(state.CloudStack.Zone).To(Equal("some-zone"))
							Expect(state.CloudStack.Endpoint).To(Equal("http://my-cloudstack.com/client/api"))
							Expect(state.CloudStack.Secure).To(BeTrue())
							Expect(state.CloudStack.IsoSegment).To(BeTrue())
						})

						It("returns the command", func() {
							appConfig, err := c.Bootstrap(bootstrapArgs(args))
							Expect(err).NotTo(HaveOccurred())

							Expect(appConfig.Command).To(Equal("up"))
						})
					})
				})

				Context("when a previous state exists", func() {
					BeforeEach(func() {
						fakeStateMigrator.MigrateCall.Returns.State = storage.State{
							IAAS: "cloudstack",
							CloudStack: storage.CloudStack{
								ApiKey:          "some-api-key",
								SecretAccessKey: "some-secret-access-key",
								Zone:            "some-zone",
								Endpoint:        "http://my-cloudstack.com/client/api",
							},
							EnvID: "some-env-id",
						}
					})

					Context("when valid matching configuration is passed in", func() {
						It("returns state with existing configuration", func() {
							appConfig, err := c.Bootstrap(bootstrapArgs([]string{
								"bbl", "up",
								"--iaas", "cloudstack",
								"--cloudstack-api-key", "some-api-key",
								"--cloudstack-secret-access-key", "some-secret-access-key",
								"--cloudstack-zone", "some-zone",
								"--cloudstack-endpoint", "http://my-cloudstack.com/client/api",
							}))
							Expect(err).NotTo(HaveOccurred())

							Expect(appConfig.State.EnvID).To(Equal("some-env-id"))
						})
					})

					DescribeTable("when non-matching configuration is passed in",
						func(args []string, expected string) {
							_, err := c.Bootstrap(bootstrapArgs(args))

							Expect(err).To(MatchError(expected))
						},
						Entry("returns an error for non-matching IAAS", []string{"bbl", "up", "--iaas", "gcp"},
							"The iaas type cannot be changed for an existing environment. The current iaas type is cloudstack."),
					)
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
						appConfig, err := c.Bootstrap(bootstrapArgs([]string{
							"bbl", "up",
						}))
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.State.EnvID).To(Equal("some-env-id"))

						workingDir, err := os.Getwd()
						Expect(err).NotTo(HaveOccurred())

						Expect(fakeStateBootstrap.GetStateCall.Receives.Dir).To(Equal(workingDir))
					})
				})

				Context("when valid matching configuration is passed in", func() {
					It("returns state with existing configuration", func() {
						appConfig, err := c.Bootstrap(bootstrapArgs([]string{
							"bbl", "up",
							"--iaas", "azure",
							"--azure-client-id", "client-id",
							"--azure-client-secret", "client-secret",
							"--azure-region", "region",
							"--azure-subscription-id", "subscription-id",
							"--azure-tenant-id", "tenant-id",
						}))
						Expect(err).NotTo(HaveOccurred())

						Expect(appConfig.State.EnvID).To(Equal("some-env-id"))
					})
				})

				DescribeTable("when non-matching configuration is passed in",
					func(args []string, expected string) {
						_, err := c.Bootstrap(bootstrapArgs(args))

						Expect(err).To(MatchError(expected))
					},
					Entry("returns an error for non-matching IAAS", []string{"bbl", "up", "--iaas", "aws"},
						"The iaas type cannot be changed for an existing environment. The current iaas type is azure."),
				)
			})
		})

		Context("when the updated, migrated configuration is invalid", func() {
			var fakeMerger *fakes.Merger
			BeforeEach(func() {
				fakeMerger = &fakes.Merger{}
				c = config.NewConfig(fakeStateBootstrap, fakeStateMigrator, fakeMerger, fakeDownloader, fakeLogger, fakeFileIO)

				fakeMerger.MergeCall.Returns.State = storage.State{
					IAAS:  "gcp",
					GCP:   storage.GCP{}, // not enough config, validate should error
					EnvID: "some-env-id",
				}
			})
			Context("and the command modifies state", func() {
				It("errors", func() {
					_, err := c.Bootstrap(bootstrapArgs([]string{
						"bbl", "up",
						"--iaas", "gcp",
					}))
					Expect(err.Error()).To(ContainSubstring("gcp-service-account-key"))
				})
			})

			Context("and the command doesn't modify state", func() {
				It("does not error", func() {
					_, err := c.Bootstrap(bootstrapArgs([]string{
						"bbl", "print-env",
					}))
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("ValidateIAAS", func() {
		DescribeTable("when configuration is invalid",
			func(state storage.State, expectedErr string) {
				err := config.ValidateIAAS(state)
				Expect(err).To(MatchError(fmt.Sprintf("\n\n%s\n", expectedErr)))
			},
			Entry("when IAAS is missing",
				storage.State{},
				"--iaas [gcp, aws, azure, vsphere, openstack, cloudstack] must be provided or BBL_IAAS must be set"),
			Entry("when IAAS is unsupported",
				storage.State{
					IAAS: "not-a-real-iaas",
				},
				"--iaas [gcp, aws, azure, vsphere, openstack, cloudstack] must be provided or BBL_IAAS must be set"),
			Entry("when an AWS credential is missing",
				storage.State{
					IAAS: "aws",
					AWS: storage.AWS{
						SecretAccessKey: "value",
						Region:          "value",
					},
				},
				"Missing --aws-access-key-id. To see all required credentials run `bbl plan --help`."),
			Entry("when a GCP credential is missing",
				storage.State{
					IAAS: "gcp",
					GCP: storage.GCP{
						ServiceAccountKey: "value",
					},
				},
				"Missing --gcp-region. To see all required credentials run `bbl plan --help`."),
			Entry("when an Azure credential is missing",
				storage.State{
					IAAS: "azure",
					Azure: storage.Azure{
						Region:         "value",
						TenantID:       "value",
						SubscriptionID: "value",
						ClientSecret:   "value",
					},
				},
				"Missing --azure-client-id. To see all required credentials run `bbl plan --help`."),
			Entry("when a vSphere credential is missing",
				storage.State{
					IAAS: "vsphere",
					VSphere: storage.VSphere{
						VCenterUser:     "value",
						VCenterPassword: "value",
						VCenterIP:       "value",
						VCenterDC:       "value",
						VCenterRP:       "value",
						VCenterDS:       "value",
						VCenterCluster:  "value",
						Network:         "value",
					},
				},
				"Missing --vsphere-subnet-cidr. To see all required credentials run `bbl plan --help`."),
			Entry("when any OpenStack credential is missing",
				storage.State{
					IAAS: "openstack",
					OpenStack: storage.OpenStack{
						AuthURL:     "value",
						AZ:          "value",
						NetworkID:   "value",
						NetworkName: "value",
						Password:    "value",
						Username:    "value",
						Project:     "value",
						Domain:      "value",
					},
				},
				"Missing --openstack-region. To see all required credentials run `bbl plan --help`."),
			Entry("when any CloudStack credential is missing",
				storage.State{
					IAAS: "cloudstack",
					CloudStack: storage.CloudStack{
						InternalCidr:    "value",
						ExternalIP:      "value",
						ApiKey:          "value",
						SecretAccessKey: "value",
						IsoSegment:      true,
						Secure:          true,
						Subnet:          "value",
					},
				},
				"Missing --cloudstack-endpoint. To see all required credentials run `bbl plan --help`."),
		)

	})
})
