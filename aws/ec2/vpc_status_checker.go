package ec2

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type ec2ClientProvider interface {
	GetEC2Client() Client
}

type VPCStatusChecker struct {
	ec2ClientProvider ec2ClientProvider
}

func NewVPCStatusChecker(ec2ClientProvider ec2ClientProvider) VPCStatusChecker {
	return VPCStatusChecker{
		ec2ClientProvider: ec2ClientProvider,
	}
}

func (v VPCStatusChecker) ValidateSafeToDelete(vpcID, envID string) error {
	output, err := v.ec2ClientProvider.GetEC2Client().DescribeInstances(&awsec2.DescribeInstancesInput{
		Filters: []*awsec2.Filter{{
			Name:   aws.String("vpc-id"),
			Values: []*string{aws.String(vpcID)},
		}},
	})
	if err != nil {
		return err
	}

	vms := v.flattenVMs(output.Reservations)
	if envID != "" {
		vms = v.removeOneVM(vms, fmt.Sprintf("%s-nat", envID))
	}
	vms = v.removeOneVM(vms, "NAT")
	vms = v.removeOneVM(vms, "bosh/0")

	if len(vms) > 0 {
		return fmt.Errorf("vpc %s is not safe to delete; vms still exist: [%s]", vpcID, strings.Join(vms, ", "))
	}

	return nil
}

func (v VPCStatusChecker) flattenVMs(reservations []*awsec2.Reservation) []string {
	vms := []string{}
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			vms = append(vms, v.vmName(instance))
		}
	}
	return vms
}

func (v VPCStatusChecker) vmName(instance *awsec2.Instance) string {
	name := "unnamed"

	for _, tag := range instance.Tags {
		if aws.StringValue(tag.Key) == "Name" && aws.StringValue(tag.Value) != "" {
			name = aws.StringValue(tag.Value)
		}
	}

	return name
}

func (v VPCStatusChecker) removeOneVM(vms []string, vmToRemove string) []string {
	for index, vm := range vms {
		if vm == vmToRemove {
			return append(vms[:index], vms[index+1:]...)
		}
	}

	return vms
}
