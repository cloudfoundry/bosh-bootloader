package azure

import "fmt"

type Group struct {
	client     groupsClient
	identifier string
}

// Group represents an Azure resource group.
func NewGroup(client groupsClient, name *string) Group {
	return Group{
		client:     client,
		identifier: *name,
	}
}

// Delete deletes an Azure resource group and all other Azure
// resources in the resource group.
func (g Group) Delete() error {
	_, errChan := g.client.Delete(g.identifier, nil)

	err := <-errChan
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (g Group) Name() string {
	return g.identifier
}

func (g Group) Type() string {
	return "Resource Group"
}
