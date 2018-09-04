package eks

import (
	"fmt"
	"strings"

	awseks "github.com/aws/aws-sdk-go/service/eks"
	"github.com/genevieve/leftovers/common"
)

type clustersClient interface {
	ListClusters(*awseks.ListClustersInput) (*awseks.ListClustersOutput, error)
	DeleteCluster(*awseks.DeleteClusterInput) (*awseks.DeleteClusterOutput, error)
}

type logger interface {
	Printf(m string, a ...interface{})
	PromptWithDetails(resourceType, resourceName string) bool
}

type Clusters struct {
	client clustersClient
	logger logger
}

func NewClusters(client clustersClient, logger logger) Clusters {
	return Clusters{
		client: client,
		logger: logger,
	}
}

func (c Clusters) List(filter string) ([]common.Deletable, error) {
	clusters, err := c.client.ListClusters(&awseks.ListClustersInput{})
	if err != nil {
		return nil, fmt.Errorf("List EKS Clusters: %s", err)
	}

	var resources []common.Deletable
	for _, cluster := range clusters.Clusters {
		r := NewCluster(c.client, cluster)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := c.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (c Clusters) Type() string {
	return "eks-cluster"
}
