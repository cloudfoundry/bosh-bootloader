package iam

import (
	"fmt"
	"strings"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
	"github.com/genevievelesperance/leftovers/aws/common"
)

type instanceProfilesClient interface {
	ListInstanceProfiles(*awsiam.ListInstanceProfilesInput) (*awsiam.ListInstanceProfilesOutput, error)
	RemoveRoleFromInstanceProfile(*awsiam.RemoveRoleFromInstanceProfileInput) (*awsiam.RemoveRoleFromInstanceProfileOutput, error)
	DeleteInstanceProfile(*awsiam.DeleteInstanceProfileInput) (*awsiam.DeleteInstanceProfileOutput, error)
}

type InstanceProfiles struct {
	client instanceProfilesClient
	logger logger
}

func NewInstanceProfiles(client instanceProfilesClient, logger logger) InstanceProfiles {
	return InstanceProfiles{
		client: client,
		logger: logger,
	}
}

func (i InstanceProfiles) List(filter string) ([]common.Deletable, error) {
	profiles, err := i.client.ListInstanceProfiles(&awsiam.ListInstanceProfilesInput{})
	if err != nil {
		return nil, fmt.Errorf("Listing instance profiles: %s", err)
	}

	var resources []common.Deletable
	for _, p := range profiles.InstanceProfiles {
		resource := NewInstanceProfile(i.client, p.InstanceProfileName, p.Roles)

		if !strings.Contains(resource.identifier, filter) {
			continue
		}

		proceed := i.logger.Prompt(fmt.Sprintf("Are you sure you want to delete instance profile %s?", resource.identifier))
		if !proceed {
			continue
		}

		for _, r := range p.Roles {
			role := *r.RoleName

			_, err := i.client.RemoveRoleFromInstanceProfile(&awsiam.RemoveRoleFromInstanceProfileInput{
				InstanceProfileName: resource.name,
				RoleName:            r.RoleName,
			})
			if err == nil {
				i.logger.Printf("SUCCESS removing role %s from instance profile %s\n", role, resource.identifier)
			} else {
				i.logger.Printf("ERROR removing role %s from instance profile %s: %s\n", role, resource.identifier, err)
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
