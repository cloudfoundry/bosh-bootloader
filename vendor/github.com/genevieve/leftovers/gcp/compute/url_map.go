package compute

import "fmt"

type UrlMap struct {
	client urlMapsClient
	name   string
	kind   string
}

func NewUrlMap(client urlMapsClient, name string) UrlMap {
	return UrlMap{
		client: client,
		name:   name,
		kind:   "url-map",
	}
}

func (u UrlMap) Delete() error {
	err := u.client.DeleteUrlMap(u.name)

	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (u UrlMap) Name() string {
	return u.name
}

func (u UrlMap) Type() string {
	return "Url Map"
}

func (u UrlMap) Kind() string {
	return u.kind
}
