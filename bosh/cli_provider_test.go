package bosh_test

import (
	"errors"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Client Provider", func() {
	var (
		allProxyGetter *fakes.AllProxyGetter
		cliProvider    bosh.CLIProvider
	)

	BeforeEach(func() {
		allProxyGetter = &fakes.AllProxyGetter{}

		cliProvider = bosh.NewCLIProvider(allProxyGetter, "some-path-to-bosh")
	})

	Describe("AuthenticatedCLI", func() {
		It("returns an authenticated bosh cli", func() {
			allProxyGetter.BoshAllProxyCall.Returns.URL = "jumpbox@some-all-proxy-url"
			cliRunner, err := cliProvider.AuthenticatedCLI(storage.Jumpbox{URL: "some-jumpbox:22"}, nil, "some-address", "some-username", "some-password", "some-fake-ca")
			Expect(err).NotTo(HaveOccurred())

			cli := cliRunner.(bosh.AuthenticatedCLI)
			Expect(cli.GlobalArgs).To(Equal([]string{
				"--environment", "some-address",
				"--client", "some-username",
				"--client-secret", "some-password",
				"--ca-cert", "some-fake-ca",
				"--non-interactive",
			}))
			Expect(cli.BOSHAllProxy).To(Equal("jumpbox@some-all-proxy-url"))
			Expect(allProxyGetter.BoshAllProxyCall.Receives.JumpboxURL).To(Equal("jumpbox@some-jumpbox:22"))
		})

		Context("when it can not get the correct key", func() {
			It("Errors", func() {
				allProxyGetter.GeneratePrivateKeyCall.Returns.Error = errors.New("fruit")
				_, err := cliProvider.AuthenticatedCLI(storage.Jumpbox{URL: "user@some-jumpbox:2222"}, nil, "some-address", "some-username", "some-password", "some-fake-ca")
				Expect(err).To(HaveOccurred())
				Expect(err).To(MatchError("fruit"))
			})
		})
	})
})
