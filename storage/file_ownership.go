package storage

import "path/filepath"

// tested and exercised via PatchDetector and GarbageCollector
var bblManaged = []string{
	"bbl-state.json",
	"create-jumpbox.sh",
	"create-director.sh",
	"delete-jumpbox.sh",
	"delete-director.sh",

	// vars
	"vars/bbl.tfvars",
	"vars/bosh-state.json",
	"vars/cloud-config-vars.yml",
	"vars/director-vars-file.yml",
	"vars/director-vars-store.yml",
	"vars/jumpbox-state.json",
	"vars/jumpbox-vars-file.yml",
	"vars/jumpbox-vars-store.yml",
	"vars/terraform.tfstate",
	"vars/terraform.tfstate.backup",

	// terraform files
	"terraform/bbl-template.tf",
	"terraform/.terraform",
	".terraform", // some versions of bbl erroneously made this terraform file

	// cloud config
	"cloud-config/cloud-config.yml",
	"cloud-config/ops.yml",

	// runtime config
	"runtime-config/runtime-config.yml",

	// directories
	"jumpbox-deployment",
	"bosh-deployment",
	"bbl-ops-files",
}

var bblManagedDirsWhichMayContainUserFiles = []string{
	"vars",
	"terraform",
	"cloud-config",
	"runtime-config",
}

// relPath must be from same dir as patterns above
func isBBLManaged(relPath string) bool {
	return matchesGlobList(relPath, bblManaged)
}

var userManaged = []string{
	"create-jumpbox-override.sh",
	"create-director-override.sh",
	"delete-jumpbox-override.sh",
	"delete-director-override.sh",

	"vars/*.tfvars",
	"terraform/*.tf",
	"cloud-config/*.yml",
}

func isUserManaged(relPath string) bool {
	return matchesGlobList(relPath, userManaged)
}

func matchesGlobList(relPath string, globList []string) bool {
	for _, pattern := range globList {
		matches, err := filepath.Match(pattern, relPath)
		if err != nil {
			panic(err) // only errors for malformed patterns
		}
		if matches {
			return true
		}
	}
	return false
}
