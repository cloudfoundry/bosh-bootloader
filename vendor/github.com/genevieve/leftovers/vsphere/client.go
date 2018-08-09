package vsphere

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
)

type Client struct {
	vmomi      *govmomi.Client
	datacenter string
}

func NewClient(vmomi *govmomi.Client, datacenter string) Client {
	return Client{
		vmomi:      vmomi,
		datacenter: datacenter,
	}
}

func (c Client) GetRootFolder(filter string) (*object.Folder, error) {
	searcher := object.NewSearchIndex(c.vmomi.Client)
	result, err := searcher.FindByInventoryPath(context.Background(), fmt.Sprintf("/%s/vm/%s", c.datacenter, filter))
	if err != nil {
		return nil, err
	}
	rootFolder, ok := result.(*object.Folder)
	if !ok {
		return nil, fmt.Errorf("expected '%#v' to be of type '*object.Folder' but it was not", result)
	}

	return rootFolder, nil
}
