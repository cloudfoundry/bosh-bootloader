package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/aws/common"
)

type securityGroupsClient interface {
	DescribeSecurityGroups(*awsec2.DescribeSecurityGroupsInput) (*awsec2.DescribeSecurityGroupsOutput, error)
	RevokeSecurityGroupIngress(*awsec2.RevokeSecurityGroupIngressInput) (*awsec2.RevokeSecurityGroupIngressOutput, error)
	RevokeSecurityGroupEgress(*awsec2.RevokeSecurityGroupEgressInput) (*awsec2.RevokeSecurityGroupEgressOutput, error)
	DeleteSecurityGroup(*awsec2.DeleteSecurityGroupInput) (*awsec2.DeleteSecurityGroupOutput, error)
}

type SecurityGroups struct {
	client securityGroupsClient
	logger logger
}

func NewSecurityGroups(client securityGroupsClient, logger logger) SecurityGroups {
	return SecurityGroups{
		client: client,
		logger: logger,
	}
}

func (e SecurityGroups) ListAll(filter string) ([]common.Deletable, error) {
	return e.get(filter)
}

func (e SecurityGroups) List(filter string) ([]common.Deletable, error) {
	resources, err := e.get(filter)
	if err != nil {
		return nil, err
	}

	var delete []common.Deletable
	for _, r := range resources {
		proceed := e.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		delete = append(delete, r)
	}

	return delete, nil
}

func (s SecurityGroups) get(filter string) ([]common.Deletable, error) {
	output, err := s.client.DescribeSecurityGroups(&awsec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describing security groups: %s", err)
	}

	var resources []common.Deletable
	for _, sg := range output.SecurityGroups {
		resource := NewSecurityGroup(s.client, sg.GroupId, sg.GroupName, sg.Tags)

		if *sg.GroupName == "default" {
			continue
		}

		if !strings.Contains(resource.Name(), filter) {
			continue
		}

		if len(sg.IpPermissions) > 0 {
			_, err := s.client.RevokeSecurityGroupIngress(&awsec2.RevokeSecurityGroupIngressInput{
				GroupId:       sg.GroupId,
				IpPermissions: sg.IpPermissions,
			})
			if err != nil {
				s.logger.Printf("ERROR revoking security group ingress for %s: %s\n", resource.Name(), err)
			}
		}

		if len(sg.IpPermissionsEgress) > 0 {
			_, err := s.client.RevokeSecurityGroupEgress(&awsec2.RevokeSecurityGroupEgressInput{
				GroupId:       sg.GroupId,
				IpPermissions: sg.IpPermissionsEgress,
			})
			if err != nil {
				s.logger.Printf("ERROR revoking security group egress for %s: %s\n", resource.Name(), err)
			}
		}

		resources = append(resources, resource)
	}

	return resources, nil
}
