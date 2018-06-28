package storage

import (
	gcpstorage "google.golang.org/api/storage/v1"
)

type client struct {
	project string

	service *gcpstorage.Service
	buckets *gcpstorage.BucketsService
	objects *gcpstorage.ObjectsService
}

func NewClient(project string, service *gcpstorage.Service) client {
	return client{
		project: project,
		service: service,
		buckets: service.Buckets,
		objects: service.Objects,
	}
}

func (c client) ListBuckets() (*gcpstorage.Buckets, error) {
	return c.buckets.List(c.project).Do()
}

func (c client) DeleteBucket(bucket string) error {
	return c.buckets.Delete(bucket).Do()
}

func (c client) ListObjects(bucket string) (*gcpstorage.Objects, error) {
	return c.objects.List(bucket).Versions(true).Do()
}

func (c client) DeleteObject(bucket, object string, generation int64) error {
	return c.objects.Delete(bucket, object).Generation(generation).Do()
}
