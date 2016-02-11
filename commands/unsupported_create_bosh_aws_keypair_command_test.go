package commands_test

import (
	"bytes"
	"crypto/rand"
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/aws/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UnsupportedCreateBoshAWSKeypairCommand", func() {
	Describe("CreateAndUploadRSAKey", func() {
		It("generates and assigns an RSA keypair to the specified AWS account", func() {
			fakeEc2 := fakes.NewEc2()

			ec2Client := aws.Ec2{
				Client: fakeEc2,
			}

			err := commands.CreateAndUploadRSAKey(rand.Reader, ec2Client)
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeEc2.ImportKeyPairCall.Receives.Name).To(ContainSubstring("keypair-"))
			Expect(fakeEc2.ImportKeyPairCall.Receives.PublicKey).To(HaveLen(381))

		})
		Context("failure cases", func() {
			It("returns an error when the RSA key could not be generated", func() {
				fakeEc2 := fakes.NewEc2()

				ec2Client := aws.Ec2{
					Client: fakeEc2,
				}
				fakeEc2.ImportKeyPairCall.Returns.Error = errors.New("ImportPublicKey failed")

				err := commands.CreateAndUploadRSAKey(bytes.NewBuffer([]byte{}), ec2Client)
				Expect(err).To(MatchError("EOF"))
			})

			It("returns an error when the public key cannot be imported", func() {
				fakeEc2 := fakes.NewEc2()

				ec2Client := aws.Ec2{
					Client: fakeEc2,
				}
				fakeEc2.ImportKeyPairCall.Returns.Error = errors.New("ImportPublicKey failed")

				err := commands.CreateAndUploadRSAKey(rand.Reader, ec2Client)
				Expect(err).To(MatchError("ImportPublicKey failed"))
			})
		})
	})
})
