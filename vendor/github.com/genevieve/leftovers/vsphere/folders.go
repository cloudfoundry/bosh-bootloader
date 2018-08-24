package vsphere

import (
	"context"
	"fmt"
	"strings"

	"github.com/genevieve/leftovers/common"
	"github.com/vmware/govmomi/object"
)

type client interface {
	GetRootFolder(filter string) (*object.Folder, error)
}

type Folders struct {
	client client
	logger logger
}

func NewFolders(client client, logger logger) Folders {
	return Folders{
		client: client,
		logger: logger,
	}
}

// List not only lists top-level folders, it also lists child and grandchild
// folders, and all VMs within those folders.
func (v Folders) List(filter string, rType string) ([]common.Deletable, error) {
	root, err := v.client.GetRootFolder(filter)
	if err != nil {
		return nil, fmt.Errorf("Getting root folder: %s", err)
	}

	return v.listChildren(root, rType)
}

func (f Folders) listChildren(parent *object.Folder, rType string) ([]common.Deletable, error) {
	var deletable []common.Deletable

	ctx := context.Background()
	children, err := parent.Children(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing children: %s", err)
	}

	for _, child := range children {
		g, ok := child.(*object.VirtualMachine)
		if ok {
			vm := NewVirtualMachine(g)

			if strings.Contains(strings.ToLower(vm.Type()), strings.ToLower(rType)) {
				proceed := f.logger.PromptWithDetails(vm.Type(), vm.Name())
				if !proceed {
					continue
				}
			} else {
				continue
			}

			deletable = append(deletable, vm)
			continue
		}

		childFolder, ok := child.(*object.Folder)
		if ok {
			grandchildren, err := f.listChildren(childFolder, rType)
			if err != nil {
				return nil, fmt.Errorf("listing grandchildren: %s", err)
			}
			deletable = append(deletable, grandchildren...)

			childFolderName, err := childFolder.Common.ObjectName(ctx)
			if err != nil {
				return nil, fmt.Errorf("Folder name: %s", err)
			}

			childFolderToDelete := NewFolder(childFolder, childFolderName)

			if strings.Contains(strings.ToLower(childFolderToDelete.Type()), strings.ToLower(rType)) {
				proceed := f.logger.PromptWithDetails(childFolderToDelete.Type(), childFolderToDelete.Name())
				if !proceed {
					continue
				}
			} else {
				continue
			}

			deletable = append(deletable, childFolderToDelete)
			continue
		}
	}

	return deletable, nil
}

func (f Folders) Type() string {
	return "folder"
}
