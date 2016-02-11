package aws_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/aws/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Ec2", func() {
	Describe("ImportPublicKey", func() {
		It("imports a public key given a name", func() {
			fakeEc2 := fakes.NewEc2()

			ec2Client := aws.Ec2{
				Client: fakeEc2,
			}

			err := ec2Client.ImportPublicKey("some-name", []byte("some-public-key"))
			Expect(err).NotTo(HaveOccurred())

			Expect(fakeEc2.ImportKeyPairCall.Receives.Name).To(Equal("some-name"))
			Expect(fakeEc2.ImportKeyPairCall.Receives.PublicKey).To(Equal([]byte("some-public-key")))
		})

		Context("failure cases", func() {
			It("returns an error when ImportKeyPair fails", func() {
				fakeEc2 := fakes.NewEc2()
				fakeEc2.ImportKeyPairCall.Returns.Error = errors.New("something bad happened")
				ec2Client := aws.Ec2{
					Client: fakeEc2,
				}

				err := ec2Client.ImportPublicKey("some-name", []byte("some-public-key"))
				Expect(err).To(MatchError("something bad happened"))
			})
		})
	})
})
