package bosh_test

import (
	"errors"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ConfigUpdater", func() {
	var (
		boshCLIProvider *fakes.BOSHCLIProvider
		boshCLI         *fakes.BOSHCLI
		configUpdater   bosh.ConfigUpdater
	)

	BeforeEach(func() {
		boshCLIProvider = &fakes.BOSHCLIProvider{}
		boshCLI = &fakes.BOSHCLI{}
		configUpdater = bosh.NewConfigUpdater(boshCLIProvider)
	})

	Describe("InitializeAuthenticatedCLI", func() {
		It("initializes an authenticated bosh cli", func() {
			expectedBOSHCLI := bosh.AuthenticatedCLI{
				BOSHExecutablePath: "some-executable-path",
			}
			boshCLIProvider.AuthenticatedCLICall.Returns.AuthenticatedCLI = expectedBOSHCLI
			state := storage.State{
				Jumpbox: storage.Jumpbox{
					URL: "some-jumpbox-url",
				},
				BOSH: storage.BOSH{
					DirectorUsername: "some-bosh-director-username",
					DirectorPassword: "some-bosh-director-password",
					DirectorAddress:  "some-bosh-director-address",
					DirectorSSLCA:    "some-bosh-director-ssl-ca",
				},
			}

			boshCLI, err := configUpdater.InitializeAuthenticatedCLI(state)
			Expect(err).NotTo(HaveOccurred())

			Expect(boshCLIProvider.AuthenticatedCLICall.CallCount).To(Equal(1))
			Expect(boshCLIProvider.AuthenticatedCLICall.Receives.Jumpbox.URL).To(Equal("some-jumpbox-url"))
			Expect(boshCLIProvider.AuthenticatedCLICall.Receives.DirectorAddress).To(Equal("some-bosh-director-address"))
			Expect(boshCLIProvider.AuthenticatedCLICall.Receives.DirectorUsername).To(Equal("some-bosh-director-username"))
			Expect(boshCLIProvider.AuthenticatedCLICall.Receives.DirectorPassword).To(Equal("some-bosh-director-password"))
			Expect(boshCLIProvider.AuthenticatedCLICall.Receives.DirectorCACert).To(Equal("some-bosh-director-ssl-ca"))
			Expect(boshCLIProvider.AuthenticatedCLICall.Receives.Stderr).To(Equal(os.Stderr))

			Expect(boshCLI).To(Equal(expectedBOSHCLI))
		})

		Context("when we can not create a bosh cli", func() {
			It("errors", func() {
				boshCLIProvider.AuthenticatedCLICall.Returns.Error = errors.New("banana")

				_, err := configUpdater.InitializeAuthenticatedCLI(storage.State{})
				Expect(err).To(MatchError("failed to create bosh cli: banana"))

			})
		})
	})

	Describe("UpdateRuntimeConfig", func() {
		It("calls the bosh cli with the correct arguments", func() {
			err := configUpdater.UpdateRuntimeConfig(boshCLI, "runtime-config-filepath", []string{"some-ops-file", "another-ops-file"}, "some-name")
			Expect(err).NotTo(HaveOccurred())

			Expect(boshCLI.RunCall.Receives.Args).To(Equal([]string{
				"update-runtime-config", "runtime-config-filepath",
				"--ops-file", "some-ops-file",
				"--ops-file", "another-ops-file",
				"--name", "some-name",
			}))
		})

	})

	Describe("UpdateCloudConfig", func() {
		It("calls the bosh cli with the correct arguments", func() {
			err := configUpdater.UpdateCloudConfig(boshCLI, "cloud-config-filepath", []string{"some-ops-file", "another-ops-file"}, "some-vars-file")
			Expect(err).NotTo(HaveOccurred())

			Expect(boshCLI.RunCall.Receives.Args).To(Equal([]string{
				"update-cloud-config", "cloud-config-filepath",
				"--ops-file", "some-ops-file",
				"--ops-file", "another-ops-file",
				"--vars-file", "some-vars-file",
			}))
		})

	})
})
