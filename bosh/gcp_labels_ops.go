package bosh

import (
	"gopkg.in/yaml.v2"
)

type gcpLabelsOp struct {
	Type  string            `yaml:"type"`
	Path  string            `yaml:"path"`
	Value map[string]string `yaml:"value"`
}

// GCPLabelsOps renders an ops file that applies GCP resource labels to the
// BOSH director and jumpbox VMs. Both deployments expose the VM cloud
// properties at /resource_pools/name=vms/cloud_properties, and the google CPI
// maps the `labels` cloud property onto GCP resource labels.
func GCPLabelsOps(labels map[string]string) ([]byte, error) {
	return yaml.Marshal([]gcpLabelsOp{
		{
			Type:  "replace",
			Path:  "/resource_pools/name=vms/cloud_properties/labels?",
			Value: labels,
		},
	})
}
