package aws

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	awslib "github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	awsroute53 "github.com/aws/aws-sdk-go/service/route53"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type EC2Client interface {
	DescribeAvailabilityZones(*awsec2.DescribeAvailabilityZonesInput) (*awsec2.DescribeAvailabilityZonesOutput, error)
	DescribeInstances(*awsec2.DescribeInstancesInput) (*awsec2.DescribeInstancesOutput, error)
	DescribeVpcs(*awsec2.DescribeVpcsInput) (*awsec2.DescribeVpcsOutput, error)
}

type Route53Client interface {
	ListHostedZonesByName(*awsroute53.ListHostedZonesByNameInput) (*awsroute53.ListHostedZonesByNameOutput, error)
}

type logger interface {
	Step(string, ...interface{})
}

type AvailabilityZones interface {
	RetrieveAZs(region string) ([]string, error)
}

type DNSZones interface {
	RetrieveDNS(domain string) string
}

type Client struct {
	ec2Client     EC2Client
	route53Client Route53Client
	logger        logger
}

func NewClient(creds storage.AWS, logger logger) Client {
	config := &awslib.Config{
		Credentials: credentials.NewStaticCredentials(creds.AccessKeyID, creds.SecretAccessKey, ""),
		Region:      awslib.String(creds.Region),
	}

	return Client{
		ec2Client:     awsec2.New(session.New(config)),
		route53Client: awsroute53.New(session.New(config)),
		logger:        logger,
	}
}

// If the parent domain for the provided url exists
// in AWS Route53, return that zone's name.
func (c Client) RetrieveDNS(url string) string {
	parentDomain := fmt.Sprintf("%s.", strings.Join(strings.Split(url, ".")[1:], "."))

	list, err := c.route53Client.ListHostedZonesByName(&awsroute53.ListHostedZonesByNameInput{
		DNSName: awslib.String(parentDomain),
	})
	if err != nil || len(list.HostedZones) == 0 {
		return ""
	}

	var found awsroute53.HostedZone
	for _, zone := range list.HostedZones {
		if *zone.Name == parentDomain {
			found = *zone
		}
	}

	if found.Id == nil {
		return ""
	}

	return parentDomain
}

// Return the AWS Availability Zones for a given region.
func (c Client) RetrieveAZs(region string) ([]string, error) {
	output, err := c.ec2Client.DescribeAvailabilityZones(&awsec2.DescribeAvailabilityZonesInput{
		Filters: []*awsec2.Filter{{
			Name:   awslib.String("region-name"),
			Values: []*string{awslib.String(region)},
		}},
	})
	if err != nil {
		return []string{}, err
	}

	azList := []string{}
	for _, az := range output.AvailabilityZones {
		if az == nil {
			return []string{}, errors.New("aws returned nil availability zone")
		}
		if az.ZoneName == nil {
			return []string{}, errors.New("aws returned availability zone with nil zone name")
		}

		azList = append(azList, *az.ZoneName)
	}

	sort.Strings(azList)

	return azList, nil
}

// Return true if the network with the provided name exists.
func (c Client) CheckExists(networkName string) (bool, error) {
	vpcs, err := c.ec2Client.DescribeVpcs(&awsec2.DescribeVpcsInput{
		Filters: []*awsec2.Filter{{
			Name: awslib.String("tag:Name"),
			Values: []*string{
				awslib.String(networkName),
			}},
		},
	})
	if err != nil {
		return false, fmt.Errorf("Failed to check vpc existence: %s", err)
	}

	if len(vpcs.Vpcs) > 0 {
		return true, nil
	}

	return false, nil
}

func (c Client) ValidateSafeToDelete(vpcID, envID string) error {
	output, err := c.ec2Client.DescribeInstances(&awsec2.DescribeInstancesInput{
		Filters: []*awsec2.Filter{{
			Name:   awslib.String("vpc-id"),
			Values: []*string{awslib.String(vpcID)},
		}},
	})
	if err != nil {
		return err
	}

	vms := c.flattenVMs(output.Reservations)
	vms = c.removeAll(vms, fmt.Sprintf("%s-nat", envID))
	vms = c.removeOneVM(vms, "NAT")
	vms = c.removeOneVM(vms, "bosh/0")
	vms = c.removeOneVM(vms, "jumpbox/0")

	if len(vms) > 0 {
		return fmt.Errorf("vpc %s is not safe to delete; vms still exist: [%s]", vpcID, strings.Join(vms, ", "))
	}

	return nil
}

func (c Client) flattenVMs(reservations []*awsec2.Reservation) []string {
	vms := []string{}
	for _, reservation := range reservations {
		for _, instance := range reservation.Instances {
			vms = append(vms, c.vmName(instance))
		}
	}
	return vms
}

func (c Client) vmName(instance *awsec2.Instance) string {
	name := "unnamed"

	for _, tag := range instance.Tags {
		if awslib.StringValue(tag.Key) == "Name" && awslib.StringValue(tag.Value) != "" {
			name = awslib.StringValue(tag.Value)
		}
	}

	return name
}

func (c Client) removeOneVM(vms []string, vmToRemove string) []string {
	for index, vm := range vms {
		if vm == vmToRemove {
			return append(vms[:index], vms[index+1:]...)
		}
	}

	return vms
}

func (c Client) removeAll(vms []string, vmToRemove string) []string {
	result := []string{}

	for _, vm := range vms {
		if vm != vmToRemove {
			result = append(result, vm)
		}
	}

	return result
}

func (c Client) GetVPC(vpcName string) (*string, error) {
	vpcs, err := c.ec2Client.DescribeVpcs(&awsec2.DescribeVpcsInput{
		Filters: []*awsec2.Filter{{
			Name:   awslib.String("tag:Name"),
			Values: []*string{awslib.String(vpcName)},
		}},
	})

	if err != nil {
		return nil, err
	}

	if len(vpcs.Vpcs) != 1 {
		return nil, fmt.Errorf("expected to receive exactly one VPC with name %s", vpcName)
	}

	return vpcs.Vpcs[0].VpcId, nil
}
