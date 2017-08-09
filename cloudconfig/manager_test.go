package cloudconfig_test

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"

	"golang.org/x/net/proxy"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Manager", func() {
	var (
		logger             *fakes.Logger
		cmd                *fakes.BOSHCommand
		opsGenerator       *fakes.CloudConfigOpsGenerator
		boshClientProvider *fakes.BOSHClientProvider
		boshClient         *fakes.BOSHClient
		socks5Proxy        *fakes.Socks5Proxy
		terraformManager   *fakes.TerraformManager
		sshKeyGetter       *fakes.SSHKeyGetter
		manager            cloudconfig.Manager

		tempDir       string
		incomingState storage.State

		baseCloudConfig []byte
	)

	BeforeEach(func() {
		logger = &fakes.Logger{}
		cmd = &fakes.BOSHCommand{}
		opsGenerator = &fakes.CloudConfigOpsGenerator{}
		boshClient = &fakes.BOSHClient{}
		boshClientProvider = &fakes.BOSHClientProvider{}
		socks5Proxy = &fakes.Socks5Proxy{}
		terraformManager = &fakes.TerraformManager{}
		sshKeyGetter = &fakes.SSHKeyGetter{}

		boshClientProvider.ClientCall.Returns.Client = boshClient

		var err error
		tempDir, err = ioutil.TempDir("", "")
		Expect(err).NotTo(HaveOccurred())

		cloudconfig.SetTempDir(func(string, string) (string, error) {
			return tempDir, nil
		})

		cmd.RunStub = func(stdout io.Writer, workingDirectory string, args []string) error {
			stdout.Write([]byte("some-cloud-config"))
			return nil
		}

		incomingState = storage.State{
			IAAS: "gcp",
			BOSH: storage.BOSH{
				DirectorAddress:  "some-director-address",
				DirectorUsername: "some-director-username",
				DirectorPassword: "some-director-password",
			},
		}

		opsGenerator.GenerateCall.Returns.OpsYAML = "some-ops"

		baseCloudConfig, err = ioutil.ReadFile("fixtures/base-cloud-config.yml")
		Expect(err).NotTo(HaveOccurred())

		manager = cloudconfig.NewManager(logger, cmd, opsGenerator, boshClientProvider, socks5Proxy, terraformManager, sshKeyGetter)
	})

	AfterEach(func() {
		cloudconfig.ResetTempDir()
	})

	Describe("Generate", func() {
		It("returns a cloud config yaml provided a valid bbl state", func() {
			expectedArgs := []string{
				"interpolate", fmt.Sprintf("%s/cloud-config.yml", tempDir),
				"-o", fmt.Sprintf("%s/ops.yml", tempDir),
			}

			cloudConfigYAML, err := manager.Generate(incomingState)
			Expect(err).NotTo(HaveOccurred())

			cloudConfig, err := ioutil.ReadFile(fmt.Sprintf("%s/cloud-config.yml", tempDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(cloudConfig).To(Equal(baseCloudConfig))

			Expect(opsGenerator.GenerateCall.Receives.State).To(Equal(incomingState))

			ops, err := ioutil.ReadFile(fmt.Sprintf("%s/ops.yml", tempDir))
			Expect(err).NotTo(HaveOccurred())
			Expect(string(ops)).To(Equal("some-ops"))

			Expect(cmd.RunCallCount()).To(Equal(1))
			_, workingDirectory, args := cmd.RunArgsForCall(0)
			Expect(workingDirectory).To(Equal(tempDir))
			Expect(args).To(Equal(expectedArgs))

			Expect(cloudConfigYAML).To(Equal("some-cloud-config"))
		})

		Context("failure cases", func() {
			Context("when temp dir fails", func() {
				BeforeEach(func() {
					cloudconfig.SetTempDir(func(string, string) (string, error) {
						return "", errors.New("failed to create temp dir")
					})
				})

				AfterEach(func() {
					cloudconfig.ResetTempDir()
				})

				It("returns an error", func() {
					_, err := manager.Generate(storage.State{})
					Expect(err).To(MatchError("failed to create temp dir"))
				})
			})

			Context("when write file fails to write cloud-config.yml", func() {
				BeforeEach(func() {
					cloudconfig.SetWriteFile(func(filename string, body []byte, mode os.FileMode) error {
						if strings.Contains(filename, "cloud-config.yml") {
							return errors.New("failed to write file")
						}
						return nil
					})
				})

				AfterEach(func() {
					cloudconfig.ResetWriteFile()
				})

				It("returns an error", func() {
					_, err := manager.Generate(storage.State{})
					Expect(err).To(MatchError("failed to write file"))
				})
			})

			Context("when ops generator fails to generate", func() {
				BeforeEach(func() {
					opsGenerator.GenerateCall.Returns.Error = errors.New("failed to generate")
				})

				It("returns an error", func() {
					_, err := manager.Generate(storage.State{})
					Expect(err).To(MatchError("failed to generate"))
				})
			})

			Context("when write file fails to write ops.yml", func() {
				BeforeEach(func() {
					cloudconfig.SetWriteFile(func(filename string, body []byte, mode os.FileMode) error {
						if strings.Contains(filename, "ops.yml") {
							return errors.New("failed to write file")
						}
						return nil
					})
				})

				AfterEach(func() {
					cloudconfig.ResetWriteFile()
				})

				It("returns an error", func() {
					_, err := manager.Generate(storage.State{})
					Expect(err).To(MatchError("failed to write file"))
				})
			})

			Context("when command fails to run", func() {
				BeforeEach(func() {
					cmd.RunReturns(errors.New("failed to run"))
				})

				It("returns an error", func() {
					_, err := manager.Generate(storage.State{})
					Expect(err).To(MatchError("failed to run"))
				})
			})
		})
	})

	Describe("Update", func() {
		Context("when no jumpbox exists", func() {
			It("logs steps taken", func() {
				err := manager.Update(incomingState)
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.StepCall.Messages).To(Equal([]string{
					"generating cloud config",
					"applying cloud config",
				}))
			})

			It("updates the bosh director with a cloud config provided a valid bbl state", func() {
				err := manager.Update(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshClientProvider.ClientCall.Receives.DirectorAddress).To(Equal("some-director-address"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorUsername).To(Equal("some-director-username"))
				Expect(boshClientProvider.ClientCall.Receives.DirectorPassword).To(Equal("some-director-password"))

				Expect(boshClient.UpdateCloudConfigCall.Receives.Yaml).To(Equal([]byte("some-cloud-config")))
			})

			Context("failure cases", func() {
				Context("when manager generate's command fails to run", func() {
					BeforeEach(func() {
						cmd.RunReturns(errors.New("failed to run"))
					})

					It("returns an error", func() {
						err := manager.Update(storage.State{})
						Expect(err).To(MatchError("failed to run"))
					})
				})

				Context("when bosh client fails to update cloud config", func() {
					BeforeEach(func() {
						boshClient.UpdateCloudConfigCall.Returns.Error = errors.New("failed to update")
					})

					It("returns an error", func() {
						err := manager.Update(storage.State{})
						Expect(err).To(MatchError("failed to update"))
					})
				})
			})
		})

		Context("when a jumpbox exists", func() {
			var (
				socks5Network string
				socks5Addr    string
				socks5Auth    *proxy.Auth
				socks5Forward proxy.Dialer
				socks5Client  *fakes.Socks5Client
			)

			BeforeEach(func() {
				incomingState.Jumpbox.Enabled = true
				terraformManager.GetOutputsCall.Returns.Outputs = map[string]interface{}{
					"jumpbox_url": "some-jumpbox-url",
				}
				sshKeyGetter.GetCall.Returns.PrivateKey = "some-private-key"

				socks5Client = &fakes.Socks5Client{}
				cloudconfig.SetProxySOCKS5(func(network, addr string, auth *proxy.Auth, forward proxy.Dialer) (proxy.Dialer, error) {
					socks5Network = network
					socks5Addr = addr
					socks5Auth = auth
					socks5Forward = forward

					return socks5Client, nil
				})
			})

			AfterEach(func() {
				cloudconfig.ResetProxySOCKS5()
			})

			It("logs steps taken", func() {
				err := manager.Update(incomingState)
				Expect(err).NotTo(HaveOccurred())
				Expect(logger.StepCall.Messages).To(Equal([]string{
					"starting socks5 proxy",
					"generating cloud config",
					"applying cloud config",
				}))
			})

			It("starts a socks5 proxy", func() {
				err := manager.Update(incomingState)
				Expect(err).NotTo(HaveOccurred())
				Expect(sshKeyGetter.GetCall.Receives.State).To(Equal(incomingState))
				Expect(terraformManager.GetOutputsCall.Receives.BBLState).To(Equal(incomingState))

				Expect(socks5Proxy.StartCall.CallCount).To(Equal(1))
				Expect(socks5Proxy.StartCall.Receives.JumpboxPrivateKey).To(Equal("some-private-key"))
				Expect(socks5Proxy.StartCall.Receives.JumpboxExternalURL).To(Equal("some-jumpbox-url"))
			})

			It("configures the bosh client", func() {
				socks5Proxy.AddrCall.Returns.Addr = "some-socks-proxy-addr"
				err := manager.Update(incomingState)
				Expect(err).NotTo(HaveOccurred())

				Expect(boshClient.ConfigureHTTPClientCall.CallCount).To(Equal(1))
				Expect(boshClient.ConfigureHTTPClientCall.Receives.Socks5Client).To(Equal(socks5Client))

				Expect(socks5Proxy.AddrCall.CallCount).To(Equal(1))

				Expect(socks5Network).To(Equal("tcp"))
				Expect(socks5Addr).To(Equal("some-socks-proxy-addr"))
				Expect(socks5Auth).To(BeNil())
				Expect(socks5Forward).To(Equal(proxy.Direct))
			})

			Context("failure cases", func() {
				It("returns an error when sshKeyGetter.Get fails", func() {
					sshKeyGetter.GetCall.Returns.Error = errors.New("failed to get jumpbox ssh key")
					err := manager.Update(incomingState)
					Expect(err).To(MatchError("failed to get jumpbox ssh key"))
				})

				It("returns an error when terraformManager.GetOutputs fails", func() {
					terraformManager.GetOutputsCall.Returns.Error = errors.New("failed to get terraform outputs")
					err := manager.Update(incomingState)
					Expect(err).To(MatchError("failed to get terraform outputs"))
				})

				It("returns an error when the socks5Proxy fails to start", func() {
					socks5Proxy.StartCall.Returns.Error = errors.New("failed to start socks5 proxy")
					err := manager.Update(incomingState)
					Expect(err).To(MatchError("failed to start socks5 proxy"))
				})

				It("returns an error when it cannot create a socks5 proxy client", func() {
					cloudconfig.SetProxySOCKS5(func(network, addr string, auth *proxy.Auth, forward proxy.Dialer) (proxy.Dialer, error) {
						return nil, errors.New("failed to create socks5 proxy client")
					})
					err := manager.Update(incomingState)
					Expect(err).To(MatchError("failed to create socks5 proxy client"))
				})
			})
		})
	})
})
