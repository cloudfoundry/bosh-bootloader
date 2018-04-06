package rds

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	awsrds "github.com/aws/aws-sdk-go/service/rds"
)

type DBCluster struct {
	client     dbClustersClient
	name       *string
	identifier string
	rtype      string
}

func NewDBCluster(client dbClustersClient, name *string) DBCluster {
	return DBCluster{
		client:     client,
		name:       name,
		identifier: *name,
		rtype:      "RDS DB Cluster",
	}
}

func (d DBCluster) Delete() error {
	_, err := d.client.DeleteDBCluster(&awsrds.DeleteDBClusterInput{
		DBClusterIdentifier: d.name,
		SkipFinalSnapshot:   aws.Bool(true),
	})

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (d DBCluster) Name() string {
	return d.identifier
}

func (d DBCluster) Type() string {
	return d.rtype
}
