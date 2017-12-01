package commands

import (
	"encoding/base64"
	"errors"
	"fmt"

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

func (l LBArgsHandler) GetLBState(iaas string, config CreateLBsConfig) (storage.LB, error) {
	var certData certs.CertData
	var err error

	if config.LBType == "" {
		return storage.LB{}, nil
	}

	// Ignore validation for Azure because it uses PFX format
	if (iaas == "azure" && config.LBType == "cf"){
		certData, err = l.certificateValidator.Read(config.CertPath, config.KeyPath, config.ChainPath)
		if err != nil {
			return storage.LB{}, fmt.Errorf("Reading certificate: %s", err)
		}

		return storage.LB{
			Type:   config.LBType,
			Cert:   base64.StdEncoding.EncodeToString(certData.Cert),
			Key:    string(certData.Key),
			Chain:  string(certData.Chain),
			Domain: config.Domain,
		}, nil
	}

	if !(iaas == "gcp" && config.LBType == "concourse") {
		certData, err = l.certificateValidator.ReadAndValidate(config.CertPath, config.KeyPath, config.ChainPath)
		if err != nil {
			return storage.LB{}, fmt.Errorf("Validate certificate: %s", err)
		}
	}

	if config.LBType == "concourse" && config.Domain != "" {
		return storage.LB{}, errors.New("domain is not implemented for concourse load balancers. Remove the --domain flag and try again.")
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
