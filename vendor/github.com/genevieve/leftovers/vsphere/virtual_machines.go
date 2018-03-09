package vsphere

import (
	"context"
	"fmt"

	"github.com/vmware/govmomi/object"
)

type VirtualMachines struct {
	client client
	logger logger
}

func NewVirtualMachines(client client, logger logger) VirtualMachines {
	return VirtualMachines{
		client: client,
		logger: logger,
	}
}

func (v VirtualMachines) List(filter string) ([]Deletable, error) {
	root, err := v.client.GetRootFolder(filter)
	if err != nil {
		return nil, fmt.Errorf("Getting root folder: %s", err)
	}

	var deletable []Deletable

	ctx := context.Background()

	children, err := root.Children(ctx)
	if err != nil {
		return nil, fmt.Errorf("Root folder children: %s", err)
	}

	for _, child := range children {
		childFolder, ok := child.(*object.Folder)
		if !ok {
			continue
		}

		grandchildren, err := childFolder.Children(ctx)
		if err != nil {
			return nil, fmt.Errorf("Folder children: %s", err)
		}

		for _, grandchild := range grandchildren {
			g, ok := grandchild.(*object.VirtualMachine)
			if !ok {
				continue
			}

			vm := NewVirtualMachine(g)

			proceed := v.logger.Prompt(fmt.Sprintf("Are you sure you want to delete virtual machine %s?", vm.Name()))
			if !proceed {
				continue
			}

			deletable = append(deletable, vm)
		}
	}

	return deletable, nil
}
