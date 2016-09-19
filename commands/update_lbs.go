package commands

import (
	"errors"
	"io/ioutil"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/flags"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

const UpdateLBsCommand = "update-lbs"

type updateLBConfig struct {
	certPath      string
	keyPath       string
	chainPath     string
	skipIfMissing bool
}

type UpdateLBs struct {
	certificateManager        certificateManager
	availabilityZoneRetriever availabilityZoneRetriever
	infrastructureManager     infrastructureManager
	awsCredentialValidator    awsCredentialValidator
	boshClientProvider        boshClientProvider
	logger                    logger
	certificateValidator      certificateValidator
	guidGenerator             guidGenerator
	stateStore                stateStore
}

func NewUpdateLBs(awsCredentialValidator awsCredentialValidator, certificateManager certificateManager,
	availabilityZoneRetriever availabilityZoneRetriever, infrastructureManager infrastructureManager, boshClientProvider boshClientProvider,
	logger logger, certificateValidator certificateValidator, guidGenerator guidGenerator, stateStore stateStore) UpdateLBs {

	return UpdateLBs{
		awsCredentialValidator:    awsCredentialValidator,
		certificateManager:        certificateManager,
		availabilityZoneRetriever: availabilityZoneRetriever,
		infrastructureManager:     infrastructureManager,
		boshClientProvider:        boshClientProvider,
		logger:                    logger,
		certificateValidator:      certificateValidator,
		guidGenerator:             guidGenerator,
		stateStore:                stateStore,
	}
}

func (c UpdateLBs) Execute(subcommandFlags []string, state storage.State) error {
	config, err := c.parseFlags(subcommandFlags)
	if err != nil {
		return err
	}

	err = c.awsCredentialValidator.Validate()
	if err != nil {
		return err
	}

	err = c.certificateValidator.Validate(UpdateLBsCommand, config.certPath, config.keyPath, config.chainPath)
	if err != nil {
		return err
	}

	if config.skipIfMissing && !lbExists(state.Stack.LBType) {
		c.logger.Println("no lb type exists, skipping...")
		return nil
	}

	if err := checkBBLAndLB(state, c.boshClientProvider, c.infrastructureManager); err != nil {
		return err
	}

	if match, err := c.checkCertificateAndChain(config.certPath, config.chainPath, state.Stack.CertificateName); err != nil {
		return err
	} else if match {
		c.logger.Println("no updates are to be performed")
		return nil
	}

	c.logger.Step("uploading new certificate")

	certificateName, err := certificateNameFor(state.Stack.LBType, c.guidGenerator, state.EnvID)
	if err != nil {
		return err
	}

	err = c.certificateManager.Create(config.certPath, config.keyPath, config.chainPath, certificateName)
	if err != nil {
		return err
	}

	if err := c.updateStack(certificateName, state.KeyPair.Name, state.Stack.Name, state.Stack.LBType, state.AWS.Region, state.EnvID); err != nil {
		return err
	}

	c.logger.Step("deleting old certificate")
	err = c.certificateManager.Delete(state.Stack.CertificateName)
	if err != nil {
		return err
	}

	state.Stack.CertificateName = certificateName

	err = c.stateStore.Set(state)
	if err != nil {
		return err
	}

	return nil
}

func (c UpdateLBs) checkCertificateAndChain(certPath string, chainPath string, oldCertName string) (bool, error) {
	localCertificate, err := ioutil.ReadFile(certPath)
	if err != nil {
		return false, err
	}

	remoteCertificate, err := c.certificateManager.Describe(oldCertName)
	if err != nil {
		return false, err
	}

	if strings.TrimSpace(string(localCertificate)) != strings.TrimSpace(remoteCertificate.Body) {
		return false, nil
	}

	if chainPath != "" {
		localChain, err := ioutil.ReadFile(chainPath)
		if err != nil {
			return false, err
		}

		if strings.TrimSpace(string(localChain)) != strings.TrimSpace(remoteCertificate.Chain) {
			return false, errors.New("you cannot change the chain after the lb has been created, please delete and re-create the lb with the chain")
		}
	}

	return true, nil
}

func (UpdateLBs) parseFlags(subcommandFlags []string) (updateLBConfig, error) {
	lbFlags := flags.New("update-lbs")

	config := updateLBConfig{}
	lbFlags.String(&config.certPath, "cert", "")
	lbFlags.String(&config.keyPath, "key", "")
	lbFlags.String(&config.chainPath, "chain", "")
	lbFlags.Bool(&config.skipIfMissing, "skip-if-missing", "", false)

	err := lbFlags.Parse(subcommandFlags)
	if err != nil {
		return config, err
	}

	return config, nil
}

func (c UpdateLBs) updateStack(certificateName string, keyPairName string, stackName string, lbType string, awsRegion, envID string) error {
	availabilityZones, err := c.availabilityZoneRetriever.Retrieve(awsRegion)
	if err != nil {
		return err
	}

	certificate, err := c.certificateManager.Describe(certificateName)
	if err != nil {
		return err
	}

	_, err = c.infrastructureManager.Update(keyPairName, len(availabilityZones), stackName, lbType, certificate.ARN, envID)
	if err != nil {
		return err
	}

	return nil
}
