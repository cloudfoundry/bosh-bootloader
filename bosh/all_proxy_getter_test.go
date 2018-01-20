package bosh_test

import (
	"errors"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("AllProxyGetter", func() {
	var (
		allProxyGetter bosh.AllProxyGetter
		sshKeyGetter   *fakes.SSHKeyGetter
		fs             *fakes.FileIO
	)

	BeforeEach(func() {
		fs = &fakes.FileIO{}
		fs.TempDirCall.Returns.Name = "some-temp-dir"

		sshKeyGetter = &fakes.SSHKeyGetter{}
		sshKeyGetter.GetCall.Returns.PrivateKey = "some-private-key"

		allProxyGetter = bosh.NewAllProxyGetter(sshKeyGetter, fs)
	})

	Describe("GeneratePrivateKey", func() {
		It("writes private key to file in temp dir", func() {
			key, err := allProxyGetter.GeneratePrivateKey()
			Expect(err).NotTo(HaveOccurred())

			Expect(key).To(Equal(filepath.Join("some-temp-dir", "bosh_jumpbox_private.key")))

			Expect(fs.WriteFileCall.Receives[0].Filename).To(Equal(key))
			Expect(fs.WriteFileCall.Receives[0].Contents).To(Equal([]byte("some-private-key")))
		})

		Context("if the ssh key getter fails to get a private key", func() {
			BeforeEach(func() {
				sshKeyGetter.GetCall.Returns.Error = errors.New("fruit")
			})

			It("returns the error", func() {
				_, err := allProxyGetter.GeneratePrivateKey()
				Expect(err).To(MatchError("fruit"))
			})
		})

		Context("when the private key can't be written", func() {
			BeforeEach(func() {
				fs.WriteFileCall.Returns = []fakes.WriteFileReturn{
					{
						Error: errors.New("mango"),
					},
				}
			})

			It("returns an error", func() {
				_, err := allProxyGetter.GeneratePrivateKey()
				Expect(err).To(MatchError("mango"))
			})
		})

	})

	Describe("BoshAllProxy", func() {
		It("interpolates a string!", func() {
			result := allProxyGetter.BoshAllProxy("jumpbox-url", "private-key-file")
			Expect(result).To(ContainSubstring("jumpbox-url"))
			Expect(result).To(ContainSubstring("private-key-file"))
		})
	})
})
