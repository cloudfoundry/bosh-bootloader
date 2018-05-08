package storage

import (
	gcpstorage "google.golang.org/api/storage/v1"
)

type client struct {
	project string
	logger  logger

	service *gcpstorage.Service
	buckets *gcpstorage.BucketsService
}

func NewClient(project string, service *gcpstorage.Service, logger logger) client {
	return client{
		project: project,
		logger:  logger,
		service: service,
		buckets: service.Buckets,
	}
}

func (c client) ListBuckets() (*gcpstorage.Buckets, error) {
	return c.buckets.List(c.project).Do()
}

func (c client) DeleteBucket(bucket string) error {
	return c.buckets.Delete(bucket).Do()
}
