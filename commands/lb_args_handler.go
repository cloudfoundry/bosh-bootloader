package commands

import (
	"errors"
	"fmt"
	"io/ioutil"

	"github.com/cloudfoundry/bosh-bootloader/certs"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type LBArgsHandler struct {
	certificateValidator certificateValidator
}

func NewLBArgsHandler(certificateValidator certificateValidator) LBArgsHandler {
	return LBArgsHandler{
		certificateValidator: certificateValidator,
	}
}

func lbExists(lbType string) bool {
	return lbType == "concourse" || lbType == "cf"
}

func ReadCerts(config CreateLBsConfig) (storage.LB, error) {
	lb := storage.LB{
		Type:   config.LBType,
		Domain: config.Domain,
	}

	certContents, err := ioutil.ReadFile(config.CertPath)
	if err != nil {
		return storage.LB{}, err
	}

	keyContents, err := ioutil.ReadFile(config.KeyPath)
	if err != nil {
		return storage.LB{}, err
	}

	lb.Cert = string(certContents)
	lb.Key = string(keyContents)

	if config.ChainPath != "" {
		chainContents, err := ioutil.ReadFile(config.ChainPath)
		if err != nil {
			return storage.LB{}, err
		}

		lb.Chain = string(chainContents)
	}

	return lb, nil
}

func (l LBArgsHandler) GetLBState(iaas string, config CreateLBsConfig) (storage.LB, error) {
	if !lbExists(config.LBType) {
		return storage.LB{}, errors.New("--type is required")
	}

	var certData certs.CertData
	var err error
	if !(iaas == "gcp" && config.LBType == "concourse") {
		certData, err = l.certificateValidator.ReadAndValidate(config.CertPath, config.KeyPath, config.ChainPath)
		if err != nil {
			return storage.LB{}, fmt.Errorf("Validate certificate: %s", err)
		}
	}

	if config.LBType == "concourse" && config.Domain != "" {
		return storage.LB{}, errors.New("--domain is not implemented for concourse load balancers. Remove the --domain flag and try again.")
	}

	return storage.LB{
		Type:   config.LBType,
		Cert:   string(certData.Cert),
		Key:    string(certData.Key),
		Chain:  string(certData.Chain),
		Domain: config.Domain,
	}, nil
}

func (l LBArgsHandler) Merge(new storage.LB, old storage.LB) storage.LB {
	if old.Type != "" {
		if new.Domain == "" {
			new.Domain = old.Domain
		}

		if new.Type == "" {
			new.Type = old.Type
		}
	}

	return new
}
