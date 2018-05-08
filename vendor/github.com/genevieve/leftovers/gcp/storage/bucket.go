package storage

import "fmt"

type Bucket struct {
	client bucketsClient
	name   string
	kind   string
}

func NewBucket(client bucketsClient, name string) Bucket {
	return Bucket{
		client: client,
		name:   name,
		kind:   "bucket",
	}
}

func (b Bucket) Delete() error {
	err := b.client.DeleteBucket(b.name)
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (b Bucket) Name() string {
	return b.name
}

func (b Bucket) Type() string {
	return "Storage Bucket"
}

func (b Bucket) Kind() string {
	return b.kind
}
