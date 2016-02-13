package ec2_test

import (
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"io"

	"github.com/pivotal-cf-experimental/bosh-bootloader/ec2"
	"golang.org/x/crypto/ssh"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeypairGenerator", func() {
	Describe("Generate", func() {
		uuidGenerator := func() (string, error) {
			return "random-uuid", nil
		}

		It("generates a valid keypair with a randomized name", func() {

			generator := ec2.NewKeypairGenerator(rand.Reader, uuidGenerator, rsa.GenerateKey, ssh.NewPublicKey)

			keypair, err := generator.Generate()
			Expect(err).NotTo(HaveOccurred())
			Expect(keypair.Name).To(Equal("keypair-random-uuid"))

			_, _, _, rest, err := ssh.ParseAuthorizedKey(keypair.Key)
			Expect(err).NotTo(HaveOccurred())
			Expect(rest).To(BeEmpty())
		})

		Context("failure cases", func() {
			Context("rsa key could not be generated", func() {
				It("returns an error", func() {
					rsaKeyGenerator := func(io.Reader, int) (*rsa.PrivateKey, error) {
						return nil, errors.New("rsa key generation failed")
					}
					generator := ec2.NewKeypairGenerator(rand.Reader, uuidGenerator, rsaKeyGenerator, ssh.NewPublicKey)

					_, err := generator.Generate()
					Expect(err).To(MatchError("rsa key generation failed"))
				})
			})

			Context("ssh public key could not be generated", func() {
				It("returns an error", func() {
					sshPublicKeyGenerator := func(interface{}) (ssh.PublicKey, error) {
						return nil, errors.New("ssh key generation failed")
					}
					generator := ec2.NewKeypairGenerator(rand.Reader, uuidGenerator, rsa.GenerateKey, sshPublicKeyGenerator)

					_, err := generator.Generate()
					Expect(err).To(MatchError("ssh key generation failed"))
				})
			})

			Context("uuid could not be generated", func() {
				It("returns an error", func() {
					uuidGenerator = func() (string, error) {
						return "", errors.New("uuid generation failed")
					}
					generator := ec2.NewKeypairGenerator(rand.Reader, uuidGenerator, rsa.GenerateKey, ssh.NewPublicKey)

					_, err := generator.Generate()
					Expect(err).To(MatchError("uuid generation failed"))
				})
			})
		})
	})
})
