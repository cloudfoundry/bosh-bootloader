package vsphere

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/find"
	"github.com/vmware/govmomi/object"
)

func DatacenterFromID(client *govmomi.Client, id string) (*object.Datacenter, error) {
	finder := find.NewFinder(client.Client, false)

	dss, err := finder.DatacenterList(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if len(dss) == 0 {
		return nil, fmt.Errorf("Couldn't find any datacenters with name \"%s\"", id)
	} else if len(dss) > 1 {
		return nil, fmt.Errorf("Found multiple datacenters with name \"%s\": %+v", id, dss)
	}

	return dss[0], nil
}
