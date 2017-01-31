package commands

import (
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

type GCPDeleteLBs struct {
	zones                   zones
	terraformOutputProvider terraformOutputProvider
	cloudConfigGenerator    gcpCloudConfigGenerator
	logger                  logger
	boshClientProvider      boshClientProvider
	stateStore              stateStore
	terraformExecutor       terraformExecutor
}

func NewGCPDeleteLBs(terraformOutputProvider terraformOutputProvider, cloudConfigGenerator gcpCloudConfigGenerator,
	zones zones, logger logger, boshClientProvider boshClientProvider, stateStore stateStore,
	terraformExecutor terraformExecutor) GCPDeleteLBs {
	return GCPDeleteLBs{
		zones: zones,
		terraformOutputProvider: terraformOutputProvider,
		cloudConfigGenerator:    cloudConfigGenerator,
		logger:                  logger,
		boshClientProvider:      boshClientProvider,
		stateStore:              stateStore,
		terraformExecutor:       terraformExecutor,
	}
}

func (g GCPDeleteLBs) Execute(state storage.State) error {
	azs := g.zones.Get(state.GCP.Region)

	terraformOutputs, err := g.terraformOutputProvider.Get(state.TFState, state.LB.Type)
	if err != nil {
		return err
	}

	g.logger.Step("generating cloud config")
	cloudConfig, err := g.cloudConfigGenerator.Generate(gcp.CloudConfigInput{
		AZs:            azs,
		Tags:           []string{terraformOutputs.InternalTag},
		NetworkName:    terraformOutputs.NetworkName,
		SubnetworkName: terraformOutputs.SubnetworkName,
	})

	boshClient := g.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername, state.BOSH.DirectorPassword)

	cloudConfigYaml, err := marshal(cloudConfig)
	if err != nil {
		return err
	}

	g.logger.Step("applying cloud config")
	err = boshClient.UpdateCloudConfig(cloudConfigYaml)
	if err != nil {
		return err
	}

	template := strings.Join([]string{terraformVarsTemplate, terraformBOSHDirectorTemplate}, "\n")

	g.logger.Step("generating terraform template")
	tfState, err := g.terraformExecutor.Apply(state.GCP.ServiceAccountKey, state.EnvID, state.GCP.ProjectID,
		state.GCP.Zone, state.GCP.Region, "", "", "", template, state.TFState)

	switch err.(type) {
	case terraform.TerraformApplyError:
		taErr := err.(terraform.TerraformApplyError)
		state.TFState = taErr.TFState()
		if setErr := g.stateStore.Set(state); setErr != nil {
			errorList := helpers.Errors{}
			errorList.Add(err)
			errorList.Add(setErr)
			return errorList
		}
		return err
	case error:
		return err
	}
	g.logger.Step("finished applying terraform template")

	state.TFState = tfState

	state.LB.Type = ""
	err = g.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}
