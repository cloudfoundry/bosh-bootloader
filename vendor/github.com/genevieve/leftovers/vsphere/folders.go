package vsphere

import (
	"context"
	"fmt"
	"strings"

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

func (v Folders) List(filter string, rType string) ([]Deletable, error) {
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

		if len(grandchildren) == 0 {
			name, err := childFolder.Common.ObjectName(ctx)
			if err != nil {
				return nil, fmt.Errorf("Folder name: %s", err)
			}

			folder := NewFolder(childFolder, name)

			if strings.Contains(strings.ToLower(folder.Type()), strings.ToLower(rType)) {
				proceed := v.logger.PromptWithDetails(folder.Type(), folder.Name())
				if !proceed {
					continue
				}
			} else {
				continue
			}

			deletable = append(deletable, folder)
		} else {
			for _, grandchild := range grandchildren {
				g, ok := grandchild.(*object.VirtualMachine)
				if !ok {
					continue
				}

				vm := NewVirtualMachine(g)

				if strings.Contains(strings.ToLower(vm.Type()), strings.ToLower(rType)) {
					proceed := v.logger.PromptWithDetails(vm.Type(), vm.Name())
					if !proceed {
						continue
					}
				} else {
					continue
				}

				deletable = append(deletable, vm)
			}
		}
	}

	return deletable, nil
}

func (f Folders) Type() string {
	return "folder"
}
