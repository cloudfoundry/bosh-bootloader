package commands

import (
	"fmt"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type Command interface {
	Execute(subcommandFlags []string, state storage.State) error
	Usage() string
}

func lbExists(lbType string) bool {
	return lbType == "concourse" || lbType == "cf"
}

func certificateNameFor(lbType string, generator guidGenerator, envid string) (string, error) {
	guid, err := generator.Generate()
	if err != nil {
		return "", err
	}

	var certificateName string

	if envid == "" {
		certificateName = fmt.Sprintf("%s-elb-cert-%s", lbType, guid)
	} else {
		certificateName = fmt.Sprintf("%s-elb-cert-%s-%s", lbType, guid, envid)
	}

	return strings.Replace(certificateName, ":", "-", -1), nil
}
