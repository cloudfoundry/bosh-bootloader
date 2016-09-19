package commands

import (
	"fmt"
	"strings"

	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type Command interface {
	Execute(subcommandFlags []string, state storage.State) error
	Usage() string
}

func bblExists(stackName string, infrastructureManager infrastructureManager, boshClient bosh.Client) error {
	if stackExists, err := infrastructureManager.Exists(stackName); err != nil {
		return err
	} else if !stackExists {
		return BBLNotFound
	}

	if _, err := boshClient.Info(); err != nil {
		return BBLNotFound
	}

	return nil
}

func checkBBLAndLB(state storage.State, boshClientProvider boshClientProvider, infrastructureManager infrastructureManager) error {
	boshClient := boshClientProvider.Client(state.BOSH.DirectorAddress, state.BOSH.DirectorUsername,
		state.BOSH.DirectorPassword)

	if err := bblExists(state.Stack.Name, infrastructureManager, boshClient); err != nil {
		return err
	}

	if !lbExists(state.Stack.LBType) {
		return LBNotFound
	}

	return nil
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
