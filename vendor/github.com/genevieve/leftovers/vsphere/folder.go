package vsphere

import (
	"context"
	"fmt"
	"time"

	"github.com/vmware/govmomi/object"
)

// Folder represents an inventory folder within vSphere.
type Folder struct {
	folder *object.Folder
	name   string
}

func NewFolder(folder *object.Folder, name string) Folder {
	return Folder{
		folder: folder,
		name:   name,
	}
}

func (f Folder) Delete() error {
	tctx, cancel := context.WithTimeout(context.Background(), time.Minute*5)
	defer cancel()

	destroy, err := f.folder.Common.Destroy(tctx)
	if err != nil {
		return fmt.Errorf("Destroy folder %s: %s", f.name, err)
	}

	err = destroy.Wait(tctx)
	if err != nil {
		return fmt.Errorf("Waiting for folder %s to destroy: %s", f.name, err)
	}

	return nil
}

func (f Folder) Name() string {
	return f.name
}

func (f Folder) Type() string {
	return "Folder"
}
