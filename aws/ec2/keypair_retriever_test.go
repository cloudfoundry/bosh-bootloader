package ec2_test

import (
	"errors"

	goaws "github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPairRetriever", func() {
	var (
		ec2Client *fakes.EC2Client
		retriever ec2.KeyPairRetriever
	)

	BeforeEach(func() {
		ec2Client = &fakes.EC2Client{}
		retriever = ec2.NewKeyPairRetriever()
	})

	Describe("Retrieve", func() {
		It("retrieves the keypair from AWS", func() {
			ec2Client.DescribeKeyPairsCall.Returns.DescribeKeyPairsOutput = &awsec2.DescribeKeyPairsOutput{
				KeyPairs: []*awsec2.KeyPairInfo{
					{
						KeyFingerprint: goaws.String("some-finger-print"),
						KeyName:        goaws.String("some-key-name"),
					},
				},
			}

			keypair, present, err := retriever.Retrieve(ec2Client, "some-key-name")
			Expect(err).NotTo(HaveOccurred())
			Expect(present).To(BeTrue())

			Expect(ec2Client.DescribeKeyPairsCall.Receives.Input).To(Equal(&awsec2.DescribeKeyPairsInput{
				KeyNames: []*string{
					goaws.String("some-key-name"),
				},
			}))

			Expect(keypair).To(Equal(ec2.KeyPairInfo{
				Fingerprint: "some-finger-print",
				Name:        "some-key-name",
			}))
		})

		It("returns a KeyPairNotFound error when the keypair does not exist", func() {
			ec2Client.DescribeKeyPairsCall.Returns.Error = errors.New("InvalidKeyPair.NotFound")

			_, present, err := retriever.Retrieve(ec2Client, "some-key-name")
			Expect(err).NotTo(HaveOccurred())
			Expect(present).To(BeFalse())
		})

		Context("failure cases", func() {
			It("returns an error when AWS communication fails", func() {
				ec2Client.DescribeKeyPairsCall.Returns.Error = errors.New("something bad happened")

				_, _, err := retriever.Retrieve(ec2Client, "some-key-name")
				Expect(err).To(MatchError("something bad happened"))
			})

			It("returns an error insufficient keypairs have been retrieved", func() {
				ec2Client.DescribeKeyPairsCall.Returns.DescribeKeyPairsOutput = &awsec2.DescribeKeyPairsOutput{
					KeyPairs: []*awsec2.KeyPairInfo{},
				}

				_, _, err := retriever.Retrieve(ec2Client, "some-key-name")
				Expect(err).To(MatchError("insufficient keypairs have been retrieved"))
			})
		})
	})
})
