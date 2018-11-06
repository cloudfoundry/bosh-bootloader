package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Merger struct {
	fs fs
}

func NewMerger(fs fs) Merger {
	return Merger{fs: fs}
}

func (m Merger) MergeGlobalFlagsToState(globalFlags GlobalFlags, state storage.State) (storage.State, error) {
	if globalFlags.IAAS != "" {
		if state.IAAS != "" && globalFlags.IAAS != state.IAAS {
			return storage.State{}, fmt.Errorf("The iaas type cannot be changed for an existing environment. The current iaas type is %s.", state.IAAS)
		}
		state.IAAS = globalFlags.IAAS
	}

	switch state.IAAS {
	case "aws":
		return m.updateAWSState(globalFlags, state)
	case "azure":
		return m.updateAzureState(globalFlags, state)
	case "gcp":
		return m.updateGCPState(globalFlags, state)
	case "vsphere":
		return m.updateVSphereState(globalFlags, state)
	case "openstack":
		return m.updateOpenStackState(globalFlags, state)
	}

	return state, nil
}

func copyFlagToState(source string, sink *string) {
	if source != "" {
		*sink = source
	}
}

func copyFlagToStateWithDefault(source string, sink *string, def string) {
	if source == "" {
		*sink = def
	} else {
		*sink = source
	}
}

func (m Merger) updateOpenStackState(globalFlags GlobalFlags, state storage.State) (storage.State, error) {
	copyFlagToState(globalFlags.OpenStackInternalCidr, &state.OpenStack.InternalCidr)
	copyFlagToState(globalFlags.OpenStackExternalIP, &state.OpenStack.ExternalIP)
	copyFlagToState(globalFlags.OpenStackAuthURL, &state.OpenStack.AuthURL)
	copyFlagToState(globalFlags.OpenStackAZ, &state.OpenStack.AZ)
	copyFlagToState(globalFlags.OpenStackDefaultKeyName, &state.OpenStack.DefaultKeyName)
	copyFlagToState(globalFlags.OpenStackDefaultSecurityGroup, &state.OpenStack.DefaultSecurityGroup)
	copyFlagToState(globalFlags.OpenStackNetworkID, &state.OpenStack.NetworkID)
	copyFlagToState(globalFlags.OpenStackPassword, &state.OpenStack.Password)
	copyFlagToState(globalFlags.OpenStackUsername, &state.OpenStack.Username)
	copyFlagToState(globalFlags.OpenStackProject, &state.OpenStack.Project)
	copyFlagToState(globalFlags.OpenStackDomain, &state.OpenStack.Domain)
	copyFlagToState(globalFlags.OpenStackRegion, &state.OpenStack.Region)
	copyFlagToState(globalFlags.OpenStackRegion, &state.OpenStack.Region)

	if globalFlags.OpenStackPrivateKey != "" {
		keyFlag := globalFlags.OpenStackPrivateKey
		if _, err := m.fs.Stat(keyFlag); err != nil {
			state.OpenStack.PrivateKey = keyFlag
		} else {
			absKeyPath, err := filepath.Abs(keyFlag)
			if err != nil {
				return storage.State{}, err
			}

			_, key, err := m.readKey(absKeyPath)
			if err != nil {
				return storage.State{}, err
			}

			state.OpenStack.PrivateKey = key
		}
	}

	return state, nil
}

func (m Merger) updateVSphereState(globalFlags GlobalFlags, state storage.State) (storage.State, error) {
	copyFlagToState(globalFlags.VSphereVCenterUser, &state.VSphere.VCenterUser)
	copyFlagToState(globalFlags.VSphereVCenterPassword, &state.VSphere.VCenterPassword)
	copyFlagToState(globalFlags.VSphereVCenterIP, &state.VSphere.VCenterIP)
	copyFlagToState(globalFlags.VSphereVCenterDC, &state.VSphere.VCenterDC)
	copyFlagToState(globalFlags.VSphereVCenterRP, &state.VSphere.VCenterRP)
	copyFlagToState(globalFlags.VSphereVCenterCluster, &state.VSphere.VCenterCluster)
	copyFlagToState(globalFlags.VSphereNetwork, &state.VSphere.Network)
	copyFlagToState(globalFlags.VSphereVCenterDS, &state.VSphere.VCenterDS)
	copyFlagToState(globalFlags.VSphereSubnetCIDR, &state.VSphere.SubnetCIDR)
	copyFlagToStateWithDefault(globalFlags.VSphereVCenterDisks, &state.VSphere.VCenterDisks, globalFlags.VSphereNetwork)
	copyFlagToStateWithDefault(globalFlags.VSphereVCenterTemplates, &state.VSphere.VCenterTemplates, fmt.Sprintf("%s_templates", globalFlags.VSphereNetwork))
	copyFlagToStateWithDefault(globalFlags.VSphereVCenterVMs, &state.VSphere.VCenterVMs, fmt.Sprintf("%s_vms", globalFlags.VSphereNetwork))

	return state, nil
}

func (m Merger) updateAWSState(globalFlags GlobalFlags, state storage.State) (storage.State, error) {
	copyFlagToState(globalFlags.AWSAccessKeyID, &state.AWS.AccessKeyID)
	copyFlagToState(globalFlags.AWSSecretAccessKey, &state.AWS.SecretAccessKey)

	if globalFlags.AWSRegion != "" {
		if state.AWS.Region != "" && globalFlags.AWSRegion != state.AWS.Region {
			return storage.State{}, fmt.Errorf("The region cannot be changed for an existing environment. The current region is %s.", state.AWS.Region)
		}
		state.AWS.Region = globalFlags.AWSRegion
	}

	return state, nil
}

func (m Merger) updateAzureState(globalFlags GlobalFlags, state storage.State) (storage.State, error) {
	copyFlagToState(globalFlags.AzureClientID, &state.Azure.ClientID)
	copyFlagToState(globalFlags.AzureClientSecret, &state.Azure.ClientSecret)
	copyFlagToState(globalFlags.AzureRegion, &state.Azure.Region)
	copyFlagToState(globalFlags.AzureSubscriptionID, &state.Azure.SubscriptionID)
	copyFlagToState(globalFlags.AzureTenantID, &state.Azure.TenantID)

	return state, nil
}

func (m Merger) updateGCPState(globalFlags GlobalFlags, state storage.State) (storage.State, error) {
	if globalFlags.GCPServiceAccountKey != "" {
		path, key, err := m.getGCPServiceAccountKey(globalFlags.GCPServiceAccountKey)
		if err != nil {
			return storage.State{}, err
		}
		state.GCP.ServiceAccountKey = key
		state.GCP.ServiceAccountKeyPath = path

		id, err := getGCPProjectID(key)
		if err != nil {
			return storage.State{}, err
		}
		if state.GCP.ProjectID != "" && id != state.GCP.ProjectID {
			return storage.State{}, fmt.Errorf("The project ID cannot be changed for an existing environment. The current project ID is %s.", state.GCP.ProjectID)
		}
		state.GCP.ProjectID = id
	}

	if globalFlags.GCPRegion != "" {
		if state.GCP.Region != "" && globalFlags.GCPRegion != state.GCP.Region {
			return storage.State{}, fmt.Errorf("The region cannot be changed for an existing environment. The current region is %s.", state.GCP.Region)
		}
		state.GCP.Region = globalFlags.GCPRegion
	}

	return state, nil
}

func (m Merger) getGCPServiceAccountKey(key string) (string, string, error) {
	if _, err := m.fs.Stat(key); err != nil {
		return m.writeGCPServiceAccountKey(key)
	}
	return m.readKey(key)
}

func (m Merger) writeGCPServiceAccountKey(contents string) (string, string, error) {
	tempFile, err := m.fs.TempFile("", "gcpServiceAccountKey.json")
	if err != nil {
		return "", "", fmt.Errorf("Creating temp file for credentials: %s", err)
	}
	err = m.fs.WriteFile(tempFile.Name(), []byte(contents), storage.StateMode)
	if err != nil {
		return "", "", fmt.Errorf("Writing credentials to temp file: %s", err)
	}
	return tempFile.Name(), contents, nil
}

func (m Merger) readKey(path string) (string, string, error) {
	keyBytes, err := m.fs.ReadFile(path)
	if err != nil {
		return "", "", fmt.Errorf("Reading key: %v", err)
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", "", fmt.Errorf("Getting absolute path to key: %v", err)
	}
	return absPath, string(keyBytes), nil
}

func getGCPProjectID(key string) (string, error) {
	p := struct {
		ProjectID string `json:"project_id"`
	}{}

	err := json.Unmarshal([]byte(key), &p)
	if err != nil {
		return "", fmt.Errorf("Unmarshalling service account key (must be valid json): %s", err)
	}

	if p.ProjectID == "" {
		return "", errors.New("Service account key is missing field `project_id`")
	}

	return p.ProjectID, nil
}
