package storage

import (
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	gcpstorage "google.golang.org/api/storage/v1"
)

type bucketsClient interface {
	ListBuckets() (*gcpstorage.Buckets, error)
	DeleteBucket(bucket string) error

	ListObjects(bucket string) (*gcpstorage.Objects, error)
	DeleteObject(bucket, object string, generation int64) error
}

type Buckets struct {
	client bucketsClient
	logger logger
}

func NewBuckets(client bucketsClient, logger logger) Buckets {
	return Buckets{
		client: client,
		logger: logger,
	}
}

func (i Buckets) List(filter string) ([]common.Deletable, error) {
	buckets, err := i.client.ListBuckets()
	if err != nil {
		return nil, fmt.Errorf("List Storage Buckets: %s", err)
	}

	var resources []common.Deletable
	for _, bucket := range buckets.Items {
		resource := NewBucket(i.client, bucket.Name)

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		proceed := i.logger.PromptWithDetails(resource.Type(), resource.Name())
		if !proceed {
			continue
		}

		resources = append(resources, resource)
	}

	return resources, nil
}

func (b Buckets) Type() string {
	return "bucket"
}
