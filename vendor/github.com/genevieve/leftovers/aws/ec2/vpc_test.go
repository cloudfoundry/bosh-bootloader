package ec2_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/ec2"
	"github.com/genevieve/leftovers/aws/ec2/fakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Vpc", func() {
	var (
		vpc          ec2.Vpc
		client       *fakes.VpcClient
		routes       *fakes.RouteTables
		subnets      *fakes.Subnets
		gateways     *fakes.InternetGateways
		resourceTags *fakes.ResourceTags
		id           *string
	)

	BeforeEach(func() {
		client = &fakes.VpcClient{}
		routes = &fakes.RouteTables{}
		subnets = &fakes.Subnets{}
		gateways = &fakes.InternetGateways{}
		resourceTags = &fakes.ResourceTags{}
		id = aws.String("the-id")
		tags := []*awsec2.Tag{}

		vpc = ec2.NewVpc(client, routes, subnets, gateways, resourceTags, id, tags)
	})

	Describe("Delete", func() {
		It("deletes the vpc", func() {
			err := vpc.Delete()
			Expect(err).NotTo(HaveOccurred())

			Expect(routes.DeleteCall.CallCount).To(Equal(1))
			Expect(routes.DeleteCall.Receives.VpcId).To(Equal(*id))

			Expect(subnets.DeleteCall.CallCount).To(Equal(1))
			Expect(subnets.DeleteCall.Receives.VpcId).To(Equal(*id))

			Expect(gateways.DeleteCall.CallCount).To(Equal(1))
			Expect(gateways.DeleteCall.Receives.VpcId).To(Equal(*id))

			Expect(client.DeleteVpcCall.CallCount).To(Equal(1))
			Expect(client.DeleteVpcCall.Receives.Input.VpcId).To(Equal(id))

			Expect(resourceTags.DeleteCall.CallCount).To(Equal(1))
			Expect(resourceTags.DeleteCall.Receives.ResourceType).To(Equal("vpc"))
			Expect(resourceTags.DeleteCall.Receives.ResourceId).To(Equal("the-id"))
		})

		Context("when deleting routes fails", func() {
			BeforeEach(func() {
				routes.DeleteCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := vpc.Delete()
				Expect(err).To(MatchError("Delete routes: banana"))
			})
		})

		Context("when deleting subnets fails", func() {
			BeforeEach(func() {
				subnets.DeleteCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := vpc.Delete()
				Expect(err).To(MatchError("Delete subnets: banana"))
			})
		})

		Context("when deleting gateways fails", func() {
			BeforeEach(func() {
				gateways.DeleteCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := vpc.Delete()
				Expect(err).To(MatchError("Delete internet gateways: banana"))
			})
		})

		Context("when deleting resource tags fails", func() {
			BeforeEach(func() {
				resourceTags.DeleteCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := vpc.Delete()
				Expect(err).To(MatchError("Delete resource tags: banana"))
			})
		})

		Context("the client fails", func() {
			BeforeEach(func() {
				client.DeleteVpcCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := vpc.Delete()
				Expect(err).To(MatchError("Delete: banana"))
			})
		})
	})

	Describe("Name", func() {
		It("returns the identifier", func() {
			Expect(vpc.Name()).To(Equal("the-id"))
		})
	})

	Describe("Type", func() {
		It("returns the type", func() {
			Expect(vpc.Type()).To(Equal("EC2 VPC"))
		})
	})
})
