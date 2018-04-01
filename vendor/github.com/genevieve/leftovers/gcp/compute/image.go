package compute

import "fmt"

type Image struct {
	client imagesClient
	name   string
}

func NewImage(client imagesClient, name string) Image {
	return Image{
		client: client,
		name:   name,
	}
}

func (i Image) Delete() error {
	err := i.client.DeleteImage(i.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (i Image) Name() string {
	return i.name
}

func (i Image) Type() string {
	return "Image"
}
