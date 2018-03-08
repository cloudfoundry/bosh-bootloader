package vsphere

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
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
	dc, err := DatacenterFromID(c.vmomi, c.datacenter)
	if err != nil {
		return nil, fmt.Errorf("Error getting datacenter from id: %s", err)
	}

	finder := find.NewFinder(c.vmomi.Client, true)

	finder.SetDatacenter(dc)

	return finder.Folder(context.Background(), filter)
}
