package commands

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type GCPCreateLBs struct {
	terraformExecutor    terraformExecutor
	terraformOutputter   terraformOutputter
	cloudConfigGenerator gcpCloudConfigGenerator
	boshClientProvider   boshClientProvider
	zones                zones
	stateStore           stateStore
}

type GCPCreateLBsConfig struct {
	LBType string
}

func NewGCPCreateLBs(terraformExecutor terraformExecutor, terraformOutputter terraformOutputter,
	cloudConfigGenerator gcpCloudConfigGenerator, boshClientProvider boshClientProvider, zones zones,
	stateStore stateStore) GCPCreateLBs {
	return GCPCreateLBs{
		terraformExecutor:    terraformExecutor,
		terraformOutputter:   terraformOutputter,
		cloudConfigGenerator: cloudConfigGenerator,
		boshClientProvider:   boshClientProvider,
		zones:                zones,
		stateStore:           stateStore,
	}
}

func (c GCPCreateLBs) Execute(config GCPCreateLBsConfig, state storage.State) error {
	if err := c.checkFastFails(config, state); err != nil {
		return err
	}

	var err error
	templateWithLB := strings.Join([]string{terraformVarsTemplate, terraformBOSHDirectorTemplate, terraformConcourseLBTemplate}, "\n\n")
	if state.TFState, err = c.terraformExecutor.Apply(state.GCP.ServiceAccountKey, state.EnvID, state.GCP.ProjectID, state.GCP.Zone,
		state.GCP.Region, templateWithLB, state.TFState); err != nil {
		return err
	}

	if err := c.stateStore.Set(state); err != nil {
		return err
	}

	network, err := c.terraformOutputter.Get(state.TFState, "network_name")
	if err != nil {
		return err
	}

	subnetwork, err := c.terraformOutputter.Get(state.TFState, "subnetwork_name")
	if err != nil {
		return err
	}

	internalTag, err := c.terraformOutputter.Get(state.TFState, "internal_tag_name")
	if err != nil {
		return err
	}

	concourseTargetPool, err := c.terraformOutputter.Get(state.TFState, "concourse_target_pool")
	if err != nil {
		return err
	}

	cloudConfig, err := c.cloudConfigGenerator.Generate(gcp.CloudConfigInput{
		AZs:            c.zones.Get(state.GCP.Region),
		Tags:           []string{internalTag},
		NetworkName:    network,
		SubnetworkName: subnetwork,
		LoadBalancer:   concourseTargetPool,
	})
	if err != nil {
		return err
	}

	manifestYAML, err := marshal(cloudConfig)
	if err != nil {
		return err
	}

	boshClient := c.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername,
		state.BOSH.DirectorPassword)

	if err := boshClient.UpdateCloudConfig(manifestYAML); err != nil {
		return err
	}

	return nil
}

func (GCPCreateLBs) checkFastFails(config GCPCreateLBsConfig, state storage.State) error {
	if config.LBType != "concourse" {
		return fmt.Errorf("%q is not a valid lb type, valid lb types are: concourse", config.LBType)
	}

	if state.IAAS != "gcp" {
		return fmt.Errorf("iaas type must be gcp")
	}
	return nil
}
