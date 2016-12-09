package commands

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
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
	logger               logger
}

type GCPCreateLBsConfig struct {
	LBType       string
	SkipIfExists bool
}

func NewGCPCreateLBs(terraformExecutor terraformExecutor, terraformOutputter terraformOutputter,
	cloudConfigGenerator gcpCloudConfigGenerator, boshClientProvider boshClientProvider, zones zones,
	stateStore stateStore, logger logger) GCPCreateLBs {
	return GCPCreateLBs{
		terraformExecutor:    terraformExecutor,
		terraformOutputter:   terraformOutputter,
		cloudConfigGenerator: cloudConfigGenerator,
		boshClientProvider:   boshClientProvider,
		zones:                zones,
		stateStore:           stateStore,
		logger:               logger,
	}
}

func (c GCPCreateLBs) Execute(config GCPCreateLBsConfig, state storage.State) error {
	boshClient := c.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername,
		state.BOSH.DirectorPassword)

	if err := c.checkFastFails(config, state, boshClient); err != nil {
		return err
	}

	if config.SkipIfExists && config.LBType == state.Stack.LBType {
		c.logger.Step(fmt.Sprintf("lb type %q exists, skipping...", config.LBType))
		return nil
	}

	c.logger.Step("generating terraform template")
	var err error
	templateWithLB := strings.Join([]string{terraformVarsTemplate, terraformBOSHDirectorTemplate, terraformConcourseLBTemplate}, "\n\n")
	if state.TFState, err = c.terraformExecutor.Apply(state.GCP.ServiceAccountKey, state.EnvID, state.GCP.ProjectID, state.GCP.Zone,
		state.GCP.Region, templateWithLB, state.TFState); err != nil {
		return err
	}
	c.logger.Step("finished applying terraform template")

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

	c.logger.Step("generating cloud config")
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

	c.logger.Step("applying cloud config")
	if err := boshClient.UpdateCloudConfig(manifestYAML); err != nil {
		return err
	}

	state.Stack.LBType = config.LBType
	if err := c.stateStore.Set(state); err != nil {
		return err
	}

	return nil
}

func (GCPCreateLBs) checkFastFails(config GCPCreateLBsConfig, state storage.State, boshClient bosh.Client) error {
	if config.LBType != "concourse" {
		return fmt.Errorf("%q is not a valid lb type, valid lb types are: concourse", config.LBType)
	}

	if state.IAAS != "gcp" {
		return fmt.Errorf("iaas type must be gcp")
	}

	_, err := boshClient.Info()
	return err
}
