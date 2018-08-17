package ec2

import (
	"fmt"
	"strings"

	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	"github.com/genevieve/leftovers/common"
)

type securityGroupsClient interface {
	DescribeSecurityGroups(*awsec2.DescribeSecurityGroupsInput) (*awsec2.DescribeSecurityGroupsOutput, error)
	RevokeSecurityGroupIngress(*awsec2.RevokeSecurityGroupIngressInput) (*awsec2.RevokeSecurityGroupIngressOutput, error)
	RevokeSecurityGroupEgress(*awsec2.RevokeSecurityGroupEgressInput) (*awsec2.RevokeSecurityGroupEgressOutput, error)
	DeleteSecurityGroup(*awsec2.DeleteSecurityGroupInput) (*awsec2.DeleteSecurityGroupOutput, error)
}

type SecurityGroups struct {
	client       securityGroupsClient
	logger       logger
	resourceTags resourceTags
}

func NewSecurityGroups(client securityGroupsClient, logger logger, resourceTags resourceTags) SecurityGroups {
	return SecurityGroups{
		client:       client,
		logger:       logger,
		resourceTags: resourceTags,
	}
}

func (s SecurityGroups) List(filter string) ([]common.Deletable, error) {
	output, err := s.client.DescribeSecurityGroups(&awsec2.DescribeSecurityGroupsInput{})
	if err != nil {
		return nil, fmt.Errorf("Describe EC2 Security Groups: %s", err)
	}

	var resources []common.Deletable
	for _, sg := range output.SecurityGroups {
		if *sg.GroupName == "default" {
			continue
		}

		r := NewSecurityGroup(s.client, s.logger, s.resourceTags, sg.GroupId, sg.GroupName, sg.Tags, sg.IpPermissions, sg.IpPermissionsEgress)

		if !strings.Contains(r.Name(), filter) {
			continue
		}

		proceed := s.logger.PromptWithDetails(r.Type(), r.Name())
		if !proceed {
			continue
		}

		resources = append(resources, r)
	}

	return resources, nil
}

func (s SecurityGroups) Type() string {
	return "ec2-security-group"
}
