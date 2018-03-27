package storage

import (
	"fmt"
	"os"
	"path/filepath"
)

var bblManaged = map[string]struct{}{
	"bbl.tfvars":               struct{}{},
	"bosh-state.json":          struct{}{},
	"cloud-config-vars.yml":    struct{}{},
	"director-vars-file.yml":   struct{}{},
	"director-vars-store.yml":  struct{}{},
	"jumpbox-state.json":       struct{}{},
	"jumpbox-vars-file.yml":    struct{}{},
	"jumpbox-vars-store.yml":   struct{}{},
	"terraform.tfstate":        struct{}{},
	"terraform.tfstate.backup": struct{}{},
}

type GarbageCollector struct {
	fs fs
}

func NewGarbageCollector(fs fs) GarbageCollector {
	return GarbageCollector{
		fs: fs,
	}
}

func (g GarbageCollector) Remove(dir string) error {
	bblStateJson := filepath.Join(dir, StateFileName)
	err := g.fs.Remove(bblStateJson)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("Removing %s: %s", bblStateJson, err)
	}

	g.fs.RemoveAll(filepath.Join(dir, "bosh-deployment"))
	g.fs.RemoveAll(filepath.Join(dir, "jumpbox-deployment"))
	g.fs.RemoveAll(filepath.Join(dir, "bbl-ops-files"))

	tfDir := filepath.Join(dir, "terraform")
	g.fs.Remove(filepath.Join(tfDir, "bbl-template.tf"))
	g.fs.RemoveAll(filepath.Join(tfDir, ".terraform"))
	g.fs.Remove(tfDir)

	ccDir := filepath.Join(dir, "cloud-config")
	g.fs.Remove(filepath.Join(ccDir, "cloud-config.yml"))
	g.fs.Remove(filepath.Join(ccDir, "ops.yml"))
	g.fs.Remove(ccDir)

	vDir := filepath.Join(dir, "vars")
	vFiles, _ := g.fs.ReadDir(vDir)
	for _, f := range vFiles {
		if _, ok := bblManaged[f.Name()]; ok {
			_ = g.fs.Remove(filepath.Join(vDir, f.Name()))
		}
	}
	g.fs.Remove(filepath.Join(dir, "vars"))

	g.fs.RemoveAll(filepath.Join(dir, ".terraform"))
	g.fs.Remove(filepath.Join(dir, "create-jumpbox.sh"))
	g.fs.Remove(filepath.Join(dir, "create-director.sh"))
	g.fs.Remove(filepath.Join(dir, "delete-jumpbox.sh"))
	g.fs.Remove(filepath.Join(dir, "delete-director.sh"))

	return nil
}
