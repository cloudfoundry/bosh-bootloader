package azure

import "fmt"

type Group struct {
	client     groupsClient
	identifier string
}

func NewGroup(client groupsClient, name *string) Group {
	return Group{
		client:     client,
		identifier: *name,
	}
}

func (g Group) Delete() error {
	_, errChan := g.client.Delete(g.identifier, nil)

	err := <-errChan

	if err != nil {
		return fmt.Errorf("FAILED deleting resource group %s: %s", g.identifier, err)
	}

	return nil
}

func (g Group) Name() string {
	return g.identifier
}

func (g Group) Type() string {
	return "resource group"
}
