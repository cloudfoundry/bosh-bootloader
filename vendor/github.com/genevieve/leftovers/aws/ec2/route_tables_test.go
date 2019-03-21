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

var _ = Describe("RouteTables", func() {
	var (
		client       *fakes.RouteTablesClient
		logger       *fakes.Logger
		resourceTags *fakes.ResourceTags

		routeTables ec2.RouteTables
	)

	BeforeEach(func() {
		client = &fakes.RouteTablesClient{}
		logger = &fakes.Logger{}
		resourceTags = &fakes.ResourceTags{}

		routeTables = ec2.NewRouteTables(client, logger, resourceTags)
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			client.DescribeRouteTablesCall.Returns.Output = &awsec2.DescribeRouteTablesOutput{
				RouteTables: []*awsec2.RouteTable{{
					RouteTableId: aws.String("the-route-table-id"),
					VpcId:        aws.String("the-vpc-id"),
				}},
			}
		})

		It("detaches and deletes the route tables", func() {
			err := routeTables.Delete("the-vpc-id")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.DescribeRouteTablesCall.CallCount).To(Equal(1))
			Expect(client.DescribeRouteTablesCall.Receives.Input.Filters[0].Name).To(Equal(aws.String("vpc-id")))
			Expect(client.DescribeRouteTablesCall.Receives.Input.Filters[0].Values[0]).To(Equal(aws.String("the-vpc-id")))
			Expect(client.DescribeRouteTablesCall.Receives.Input.Filters[1].Name).To(Equal(aws.String("association.main")))
			Expect(client.DescribeRouteTablesCall.Receives.Input.Filters[1].Values[0]).To(Equal(aws.String("false")))

			Expect(client.DeleteRouteTableCall.CallCount).To(Equal(1))
			Expect(client.DeleteRouteTableCall.Receives.Input.RouteTableId).To(Equal(aws.String("the-route-table-id")))

			Expect(resourceTags.DeleteCall.CallCount).To(Equal(1))
			Expect(resourceTags.DeleteCall.Receives.ResourceType).To(Equal("route-table"))
			Expect(resourceTags.DeleteCall.Receives.ResourceId).To(Equal("the-route-table-id"))

			Expect(logger.PrintfCall.Messages).To(Equal([]string{
				"[EC2 VPC: the-vpc-id] Deleted route table the-route-table-id \n",
				"[EC2 VPC: the-vpc-id] Deleted route table the-route-table-id tags \n",
			}))
		})

		Context("when the client fails to describe route tables", func() {
			BeforeEach(func() {
				client.DescribeRouteTablesCall.Returns.Error = errors.New("some error")
			})

			It("returns the error and does not try deleting them", func() {
				err := routeTables.Delete("banana")
				Expect(err).To(MatchError("Describe EC2 Route Tables: some error"))

				Expect(client.DeleteRouteTableCall.CallCount).To(Equal(0))
			})
		})

		Context("when the route table has an association id", func() {
			BeforeEach(func() {
				client.DescribeRouteTablesCall.Returns.Output = &awsec2.DescribeRouteTablesOutput{
					RouteTables: []*awsec2.RouteTable{{
						RouteTableId: aws.String("the-route-table-id"),
						VpcId:        aws.String("the-vpc-id"),
						Associations: []*awsec2.RouteTableAssociation{{
							Main:                    aws.Bool(false),
							RouteTableAssociationId: aws.String("the-association-id"),
							RouteTableId:            aws.String("the-route-table-id"),
							SubnetId:                aws.String("the-subnet-id"),
						}},
					}},
				}
			})

			It("disassociates it from the subnet before trying to delete it", func() {
				err := routeTables.Delete("the-vpc-id")
				Expect(err).NotTo(HaveOccurred())

				Expect(client.DescribeRouteTablesCall.CallCount).To(Equal(1))
				Expect(client.DisassociateRouteTableCall.CallCount).To(Equal(1))
				Expect(client.DisassociateRouteTableCall.Receives.Input.AssociationId).To(Equal(aws.String("the-association-id")))
				Expect(client.DeleteRouteTableCall.CallCount).To(Equal(1))

				Expect(logger.PrintfCall.Messages).To(Equal([]string{
					"[EC2 VPC: the-vpc-id] Disassociated route table the-route-table-id \n",
					"[EC2 VPC: the-vpc-id] Deleted route table the-route-table-id \n",
					"[EC2 VPC: the-vpc-id] Deleted route table the-route-table-id tags \n",
				}))
			})

			Context("when the client fails to disassociate the route table", func() {
				BeforeEach(func() {
					client.DisassociateRouteTableCall.Returns.Error = errors.New("some error")
				})

				It("logs the error", func() {
					err := routeTables.Delete("the-vpc-id")
					Expect(err).NotTo(HaveOccurred())

					Expect(client.DisassociateRouteTableCall.CallCount).To(Equal(1))
					Expect(logger.PrintfCall.Messages).To(Equal([]string{
						"[EC2 VPC: the-vpc-id] Disassociate route table the-route-table-id: some error \n",
						"[EC2 VPC: the-vpc-id] Deleted route table the-route-table-id \n",
						"[EC2 VPC: the-vpc-id] Deleted route table the-route-table-id tags \n",
					}))
				})
			})
		})

		Context("when the client fails to delete the route table", func() {
			BeforeEach(func() {
				client.DeleteRouteTableCall.Returns.Error = errors.New("banana")
			})

			It("returns the error", func() {
				err := routeTables.Delete("the-vpc-id")
				Expect(err).To(MatchError("Delete the-route-table-id: banana"))
			})
		})
	})
})
