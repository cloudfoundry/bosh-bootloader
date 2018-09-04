package eks_test

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	awseks "github.com/aws/aws-sdk-go/service/eks"
	"github.com/genevieve/leftovers/aws/eks"
	"github.com/genevieve/leftovers/aws/eks/fakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Clusters", func() {
	var (
		client *fakes.ClustersClient
		logger *fakes.Logger

		clusters eks.Clusters
	)

	BeforeEach(func() {
		client = &fakes.ClustersClient{}
		logger = &fakes.Logger{}
		logger.PromptWithDetailsCall.Returns.Proceed = true

		clusters = eks.NewClusters(client, logger)
	})

	Describe("List", func() {
		BeforeEach(func() {
			client.ListClustersCall.Returns.Output = &awseks.ListClustersOutput{
				Clusters: []*string{aws.String("the-cluster-id")},
			}
		})

		It("returns a list of eks clusters to delete", func() {
			items, err := clusters.List("")
			Expect(err).NotTo(HaveOccurred())

			Expect(client.ListClustersCall.CallCount).To(Equal(1))

			Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
			Expect(logger.PromptWithDetailsCall.Receives.Type).To(Equal("EKS Cluster"))
			Expect(logger.PromptWithDetailsCall.Receives.Name).To(Equal("the-cluster-id"))

			Expect(items).To(HaveLen(1))
		})

		Context("when the client fails to list clusters", func() {
			BeforeEach(func() {
				client.ListClustersCall.Returns.Error = errors.New("some error")
			})

			It("returns the error", func() {
				_, err := clusters.List("")
				Expect(err).To(MatchError("List EKS Clusters: some error"))
			})
		})

		Context("when the user responds no to the prompt", func() {
			BeforeEach(func() {
				logger.PromptWithDetailsCall.Returns.Proceed = false
			})

			It("does not return it to the list", func() {
				items, err := clusters.List("")
				Expect(err).NotTo(HaveOccurred())

				Expect(logger.PromptWithDetailsCall.CallCount).To(Equal(1))
				Expect(items).To(HaveLen(0))
			})
		})
	})
})
