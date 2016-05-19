package commands

import (
	"errors"
	"io/ioutil"

	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type updateLBConfig struct {
	certPath string
	keyPath  string
}

type UpdateLBs struct {
	certificateManager        certificateManager
	availabilityZoneRetriever availabilityZoneRetriever
	infrastructureManager     infrastructureManager
	awsCredentialValidator    awsCredentialValidator
}

func NewUpdateLBs(awsCredentialValidator awsCredentialValidator, certificateManager certificateManager, availabilityZoneRetriever availabilityZoneRetriever,
	infrastructureManager infrastructureManager) UpdateLBs {

	return UpdateLBs{
		awsCredentialValidator:    awsCredentialValidator,
		certificateManager:        certificateManager,
		availabilityZoneRetriever: availabilityZoneRetriever,
		infrastructureManager:     infrastructureManager,
	}
}

func (c UpdateLBs) Execute(subcommandFlags []string, state storage.State) (storage.State, error) {
	err := c.awsCredentialValidator.Validate()
	if err != nil {
		return state, err
	}

	config, err := c.parseFlags(subcommandFlags)
	if err != nil {
		return state, err
	}

	if err := c.checkFastFails(state.Stack.Name, state.Stack.LBType); err != nil {
		return state, err
	}

	if match, err := c.certificatesMatch(config.certPath, state.Stack.CertificateName); err != nil {
		return state, err
	} else if match {
		return state, nil
	}

	certificateName, err := c.certificateManager.Create(config.certPath, config.keyPath)
	if err != nil {
		return state, err
	}

	if err := c.updateStack(certificateName, state.KeyPair.Name, state.Stack.Name, state.Stack.LBType, state.AWS.Region); err != nil {
		return state, err
	}

	err = c.certificateManager.Delete(state.Stack.CertificateName)
	if err != nil {
		return state, err
	}

	state.Stack.CertificateName = certificateName

	return state, nil
}

func (c UpdateLBs) certificatesMatch(certPath string, oldCertName string) (bool, error) {
	localCertificate, err := ioutil.ReadFile(certPath)
	if err != nil {
		return false, err
	}

	remoteCertificate, err := c.certificateManager.Describe(oldCertName)
	if err != nil {
		return false, err
	}

	return string(localCertificate) == remoteCertificate.Body, nil
}

func (UpdateLBs) parseFlags(subcommandFlags []string) (updateLBConfig, error) {
	lbFlags := flags.New("update-lbs")

	config := updateLBConfig{}
	lbFlags.String(&config.certPath, "cert", "")
	lbFlags.String(&config.keyPath, "key", "")

	err := lbFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (c UpdateLBs) updateStack(certificateName string, keyPairName string, stackName string, lbType string, awsRegion string) error {
	availabilityZones, err := c.availabilityZoneRetriever.Retrieve(awsRegion)
	if err != nil {
		return err
	}

	certificate, err := c.certificateManager.Describe(certificateName)
	if err != nil {
		return err
	}

	_, err = c.infrastructureManager.Update(keyPairName, len(availabilityZones), stackName, lbType, certificate.ARN)
	if err != nil {
		return err
	}

	return nil
}

func (c UpdateLBs) checkFastFails(stackName string, lbType string) error {
	stackExists, err := c.infrastructureManager.Exists(stackName)
	if err != nil {
		return err
	}

	if !stackExists {
		return BBLNotFound
	}

	if lbType != "concourse" && lbType != "cf" {
		return errors.New("no load balancer has been found for this bbl environment")
	}

	return nil
}
