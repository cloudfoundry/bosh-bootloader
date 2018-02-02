package s3

import (
	"fmt"
	"strings"

	awss3 "github.com/aws/aws-sdk-go/service/s3"
	"github.com/genevievelesperance/leftovers/aws/common"
)

type bucketsClient interface {
	ListBuckets(*awss3.ListBucketsInput) (*awss3.ListBucketsOutput, error)
	DeleteBucket(*awss3.DeleteBucketInput) (*awss3.DeleteBucketOutput, error)

	ListObjectVersions(*awss3.ListObjectVersionsInput) (*awss3.ListObjectVersionsOutput, error)
	DeleteObjects(*awss3.DeleteObjectsInput) (*awss3.DeleteObjectsOutput, error)
}

type Buckets struct {
	client  bucketsClient
	logger  logger
	manager bucketManager
}

func NewBuckets(client bucketsClient, logger logger, manager bucketManager) Buckets {
	return Buckets{
		client:  client,
		logger:  logger,
		manager: manager,
	}
}

func (b Buckets) List(filter string) ([]common.Deletable, error) {
	buckets, err := b.client.ListBuckets(&awss3.ListBucketsInput{})
	if err != nil {
		return nil, fmt.Errorf("Listing buckets: %s", err)
	}

	var resources []common.Deletable
	for _, bucket := range buckets.Buckets {
		resource := NewBucket(b.client, bucket.Name)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		if !b.manager.IsInRegion(resource.identifier) {
			continue
		}

		proceed := b.logger.Prompt(fmt.Sprintf("Are you sure you want to delete bucket %s?", resource.identifier))
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
