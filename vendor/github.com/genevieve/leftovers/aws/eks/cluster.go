package eks

import (
	"fmt"

	awseks "github.com/aws/aws-sdk-go/service/eks"
)

type Cluster struct {
	client clustersClient
	id     *string
	rtype  string
}

func NewCluster(client clustersClient, id *string) Cluster {
	return Cluster{
		client: client,
		id:     id,
		rtype:  "EKS Cluster",
	}
}

func (c Cluster) Delete() error {
	_, err := c.client.DeleteCluster(&awseks.DeleteClusterInput{Name: c.id})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (c Cluster) Name() string {
	return *c.id
}

func (c Cluster) Type() string {
	return c.rtype
}
