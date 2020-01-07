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

type LBArgs struct {
	LBType   string
	CertPath string
	KeyPath  string
	Domain   string
}

func NewLBArgsHandler(certificateValidator certificateValidator) LBArgsHandler {
	return LBArgsHandler{
		certificateValidator: certificateValidator,
	}
}

func (l LBArgsHandler) GetLBState(iaas string, args LBArgs) (storage.LB, error) {
	if args.LBType == "" {
		return storage.LB{}, nil
	}

	var certData certs.CertData
	var err error

	if iaas == "azure" && args.LBType == "cf" {
		certData, err = l.certificateValidator.ReadAndValidatePKCS12(args.CertPath, args.KeyPath)
		if err != nil {
			return storage.LB{}, fmt.Errorf("Validate certificate: %s", err)
		}

		return storage.LB{
			Type:   args.LBType,
			Cert:   base64.StdEncoding.EncodeToString(certData.Cert),
			Key:    string(certData.Key),
			Domain: args.Domain,
		}, nil
	}

	if args.LBType != "concourse" {
		certData, err = l.certificateValidator.ReadAndValidate(args.CertPath, args.KeyPath)
		if err != nil {
			return storage.LB{}, fmt.Errorf("Validate certificate: %s", err)
		}
	}

	if args.LBType == "concourse" && args.Domain != "" {
		return storage.LB{}, errors.New("domain is not implemented for concourse load balancers. Remove the --lb-domain flag and try again.")
	}

	return storage.LB{
		Type:   args.LBType,
		Cert:   string(certData.Cert),
		Key:    string(certData.Key),
		Domain: args.Domain,
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
