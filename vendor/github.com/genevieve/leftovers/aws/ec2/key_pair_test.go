package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("KeyPair", func() {
	var (
		keyPair ec2.KeyPair
		client  *fakes.KeyPairsClient
		name    *string
	)

	BeforeEach(func() {
		client = &fakes.KeyPairsClient{}
		name = aws.String("the-name")

		keyPair = ec2.NewKeyPair(client, name)
	})

	Describe("Delete", func() {
		It("deletes the key pair", func() {
			err := keyPair.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DeleteKeyPairCall.CallCount).To(Equal(1))
			Expect(client.DeleteKeyPairCall.Receives.Input.KeyName).To(Equal(name))
		})

		Context("the client fails", func() {
			BeforeEach(func() {
				client.DeleteKeyPairCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := keyPair.Delete()
				Expect(err).To(MatchError("FAILED deleting key pair the-name: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(keyPair.Name()).To(Equal("the-name"))
		})
	})

	Describe("Type", func() {
		It("returns \"key pair\"", func() {
			Expect(keyPair.Type()).To(Equal("key pair"))
		})
	})
})
