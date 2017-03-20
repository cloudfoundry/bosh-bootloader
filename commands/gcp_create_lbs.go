package commands

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	"github.com/cloudfoundry/multierror"
)

type zones interface {
	Get(string) []string
}

type GCPCreateLBs struct {
	terraformExecutor  terraformExecutor
	boshClientProvider boshClientProvider
	cloudConfigManager cloudConfigManager
	zones              zones
	stateStore         stateStore
	logger             logger
}

type GCPCreateLBsConfig struct {
	LBType       string
	CertPath     string
	KeyPath      string
	Domain       string
	SkipIfExists bool
}

func NewGCPCreateLBs(terraformExecutor terraformExecutor,
	boshClientProvider boshClientProvider, cloudConfigManager cloudConfigManager, zones zones,
	stateStore stateStore, logger logger) GCPCreateLBs {
	return GCPCreateLBs{
		terraformExecutor:  terraformExecutor,
		boshClientProvider: boshClientProvider,
		cloudConfigManager: cloudConfigManager,
		zones:              zones,
		stateStore:         stateStore,
		logger:             logger,
	}
}

func (c GCPCreateLBs) Execute(config GCPCreateLBsConfig, state storage.State) error {
	err := fastFailTerraformVersion(c.terraformExecutor)
	if err != nil {
		return err
	}

	if err = c.checkFastFails(config, state); err != nil {
		return err
	}

	if !state.NoDirector {
		boshClient := c.boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername,
			state.BOSH.DirectorPassword)

		if err = c.checkBOSHClient(boshClient); err != nil {
			return err
		}
	}

	if config.SkipIfExists && config.LBType == state.LB.Type {
		c.logger.Step(fmt.Sprintf("lb type %q exists, skipping...", config.LBType))
		return nil
	}

	c.logger.Step("generating terraform template")

	var lbTemplate string
	var cert, key []byte
	zones := c.zones.Get(state.GCP.Region)
	switch config.LBType {
	case "concourse":
		lbTemplate = terraformConcourseLBTemplate
	case "cf":
		terraformCFLBBackendService := generateBackendServiceTerraform(len(zones))
		instanceGroups := generateInstanceGroups(zones)
		if config.Domain != "" {
			lbTemplate = strings.Join([]string{terraformCFLBTemplate, instanceGroups, terraformCFLBBackendService, terraformCFDNSTemplate}, "\n")
		} else {
			lbTemplate = strings.Join([]string{terraformCFLBTemplate, instanceGroups, terraformCFLBBackendService}, "\n")
		}

		cert, err = ioutil.ReadFile(config.CertPath)
		if err != nil {
			return err
		}

		key, err = ioutil.ReadFile(config.KeyPath)
		if err != nil {
			return err
		}
	}

	templateWithLB := strings.Join([]string{terraform.VarsTemplate, terraformBOSHDirectorTemplate, lbTemplate}, "\n")
	tfState, err := c.terraformExecutor.Apply(state.GCP.ServiceAccountKey, state.EnvID, state.GCP.ProjectID, state.GCP.Zone,
		state.GCP.Region, string(cert), string(key), config.Domain, templateWithLB, state.TFState)
	switch err.(type) {
	case terraform.ExecutorApplyError:
		taError := err.(terraform.ExecutorApplyError)
		state.TFState = taError.TFState()
		if setErr := c.stateStore.Set(state); setErr != nil {
			errorList := helpers.Errors{}
			errorList.Add(err)
			errorList.Add(setErr)
			return errorList
		}
		return taError
	case error:
		return err
	}
	c.logger.Step("finished applying terraform template")

	state.TFState = tfState
	if err := c.stateStore.Set(state); err != nil {
		return err
	}

	state.LB.Type = config.LBType
	if config.LBType == "cf" {
		state.LB.Cert = string(cert)
		state.LB.Key = string(key)
		state.LB.Domain = config.Domain
	}

	if !state.NoDirector {
		err = c.cloudConfigManager.Update(state)
		if err != nil {
			return err
		}
	}

	if err := c.stateStore.Set(state); err != nil {
		return err
	}

	return nil
}

func (GCPCreateLBs) checkFastFails(config GCPCreateLBsConfig, state storage.State) error {
	if config.LBType == "" {
		return fmt.Errorf("--type is a required flag")
	}

	if config.LBType != "concourse" && config.LBType != "cf" {
		return fmt.Errorf("%q is not a valid lb type, valid lb types are: concourse, cf", config.LBType)
	}

	if config.LBType == "cf" {
		errs := multierror.NewMultiError("create-lbs")
		if config.CertPath == "" {
			errs.Add(errors.New("--cert is required"))
		}
		if config.KeyPath == "" {
			errs.Add(errors.New("--key is required"))
		}
		if errs.Length() > 0 {
			return errs
		}
	}

	if state.IAAS != "gcp" {
		return fmt.Errorf("iaas type must be gcp")
	}

	return nil
}

func (GCPCreateLBs) checkBOSHClient(boshClient bosh.Client) error {
	_, err := boshClient.Info()
	if err != nil {
		return BBLNotFound
	}

	return nil
}

func generateBackendServiceTerraform(count int) string {
	backendResourceStart := `resource "google_compute_backend_service" "router-lb-backend-service" {
  name        = "${var.env_id}-router-lb"
  port_name   = "http"
  protocol    = "HTTP"
  timeout_sec = 900
  enable_cdn  = false
`
	backendResourceEnd := `  health_checks = ["${google_compute_http_health_check.cf-public-health-check.self_link}"]
}
`
	backendStrings := []string{}
	for i := 0; i < count; i++ {
		backendString := fmt.Sprintf(`  backend {
    group = "${google_compute_instance_group.router-lb-%d.self_link}"
  }
`, i)
		backendStrings = append(backendStrings, backendString)
	}

	backendServiceTemplate := []string{backendResourceStart}
	backendServiceTemplate = append(backendServiceTemplate, backendStrings...)
	backendServiceTemplate = append(backendServiceTemplate, backendResourceEnd)
	return strings.Join(backendServiceTemplate, "\n")
}

func generateInstanceGroups(zones []string) string {
	var groups []string
	for i, zone := range zones {
		groups = append(groups, fmt.Sprintf(`resource "google_compute_instance_group" "router-lb-%[1]d" {
  name        = "${var.env_id}-router-%[2]s"
  description = "terraform generated instance group that is multi-zone for https loadbalancing"
  zone        = "%[2]s"
}
`, i, zone))
	}

	return strings.Join(groups, "\n")
}
