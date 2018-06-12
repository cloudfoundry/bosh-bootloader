package iam

import (
	"fmt"
	"strings"

	awsiam "github.com/aws/aws-sdk-go/service/iam"
)

type InstanceProfile struct {
	client     instanceProfilesClient
	name       *string
	identifier string
	roles      []*awsiam.Role
	logger     logger
	rtype      string
}

func NewInstanceProfile(client instanceProfilesClient, name *string, roles []*awsiam.Role, logger logger) InstanceProfile {
	identifier := *name

	extra := []string{}
	for _, r := range roles {
		extra = append(extra, fmt.Sprintf("Role:%s", *r.RoleName))
	}

	if len(extra) > 0 {
		identifier = fmt.Sprintf("%s (%s)", *name, strings.Join(extra, ", "))
	}

	return InstanceProfile{
		client:     client,
		name:       name,
		identifier: identifier,
		roles:      roles,
		logger:     logger,
		rtype:      "IAM Instance Profile",
	}
}

func (i InstanceProfile) Delete() error {
	for _, r := range i.roles {
		role := *r.RoleName

		_, err := i.client.RemoveRoleFromInstanceProfile(&awsiam.RemoveRoleFromInstanceProfileInput{
			InstanceProfileName: i.name,
			RoleName:            r.RoleName,
		})
		if err == nil {
			i.logger.Printf("[%s: %s] Removed role %s \n", i.rtype, i.identifier, role)
		} else {
			i.logger.Printf("[%s: %s] Remove role %s: %s \n", i.rtype, i.identifier, role, err)
		}
	}

	_, err := i.client.DeleteInstanceProfile(&awsiam.DeleteInstanceProfileInput{InstanceProfileName: i.name})
	if err != nil {
		return fmt.Errorf("Delete: %s", err)
	}

	return nil
}

func (i InstanceProfile) Name() string {
	return i.identifier
}

func (i InstanceProfile) Type() string {
	return i.rtype
}
