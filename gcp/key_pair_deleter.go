package gcp

import (
	"fmt"
	"strings"
)

type KeyPairDeleter struct {
	clientProvider clientProvider
	logger         logger
}

func NewKeyPairDeleter(clientProvider clientProvider, logger logger) KeyPairDeleter {
	return KeyPairDeleter{
		clientProvider: clientProvider,
		logger:         logger,
	}
}

func (k KeyPairDeleter) Delete(projectID, publicKey string) error {
	k.logger.Step("deleting keypair")

	client := k.clientProvider.Client()
	project, err := client.GetProject(projectID)
	if err != nil {
		return err
	}

	sshKey := fmt.Sprintf("vcap:ssh-rsa %s vcap", publicKey)

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

	_, err = client.SetCommonInstanceMetadata(projectID, project.CommonInstanceMetadata)
	if err != nil {
		return err
	}

	return nil
}
