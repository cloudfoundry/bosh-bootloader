package gcp

import (
	"fmt"
	"strings"
)

type KeyPairDeleter struct {
	client metadataSetter
	logger logger
}

func NewKeyPairDeleter(client metadataSetter, logger logger) KeyPairDeleter {
	return KeyPairDeleter{
		client: client,
		logger: logger,
	}
}

func (k KeyPairDeleter) Delete(publicKey string) error {
	k.logger.Step("deleting keypair")

	project, err := k.client.GetProject()
	if err != nil {
		return err
	}

	sshKey := fmt.Sprintf("vcap:%s vcap", publicKey)

	var modified bool
	for i, item := range project.CommonInstanceMetadata.Items {
		if item.Key == "sshKeys" {
			newSSHKeys := []string{}

			for _, keyFromGCP := range strings.Split(*item.Value, "\n") {
				if keyFromGCP != sshKey {
					newSSHKeys = append(newSSHKeys, keyFromGCP)
				} else {
					modified = true
				}
			}

			newValue := strings.Join(newSSHKeys, "\n")
			project.CommonInstanceMetadata.Items[i].Value = &newValue

			break
		}
	}

	if !modified {
		return nil
	}

	_, err = k.client.SetCommonInstanceMetadata(project.CommonInstanceMetadata)
	if err != nil {
		return err
	}

	return nil
}
