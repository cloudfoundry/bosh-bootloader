package storage

import "fmt"

type Bucket struct {
	client bucketsClient
	name   string
}

func NewBucket(client bucketsClient, name string) Bucket {
	return Bucket{
		client: client,
		name:   name,
	}
}

func (b Bucket) Delete() error {
	objects, err := b.client.ListObjects(b.name)
	if err != nil {
		return fmt.Errorf("List Objects: %s", err)
	}

	for _, object := range objects.Items {
		err = b.client.DeleteObject(b.name, object.Name, object.Generation)
		if err != nil {
			return fmt.Errorf("Delete Object: %s", err)
		}
	}

	err = b.client.DeleteBucket(b.name)
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
