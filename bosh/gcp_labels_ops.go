package bosh

import (
	"fmt"
	"sort"

	"gopkg.in/yaml.v2"
)

type gcpLabelsOp struct {
	Type  string      `yaml:"type"`
	Path  string      `yaml:"path"`
	Value interface{} `yaml:"value"`
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

// GCPLabelsRuntimeConfigOps renders an ops file that applies the GCP resource
// labels as runtime config tags. The director combines runtime config tags with
// the tags of each deployment, so every VM deployed to the environment is
// labelled, not only the VMs that bbl creates itself.
//
// One replace operation is emitted per label key rather than a single replace of
// the whole /tags map, so that tags defined by other ops files in the same
// runtime config are preserved. The `?` marks tags and the key as optional so
// the operations also create them when they do not exist yet.
//
// BOSH enforces its own reserved VM tags (director, deployment, instance_group,
// job, id, name, index and created_at) and will ignore a label that uses one of
// those keys on deployment VMs.
func GCPLabelsRuntimeConfigOps(labels map[string]string) ([]byte, error) {
	keys := make([]string, 0, len(labels))
	for key := range labels {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	ops := make([]gcpLabelsOp, 0, len(keys))
	for _, key := range keys {
		ops = append(ops, gcpLabelsOp{
			Type:  "replace",
			Path:  fmt.Sprintf("/tags?/%s?", key),
			Value: labels[key],
		})
	}

	return yaml.Marshal(ops)
}
