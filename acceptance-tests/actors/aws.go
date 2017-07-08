package actors

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/clientmanager"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"

	. "github.com/onsi/gomega"

	awslib "github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
)

type AWS struct {
	stackManager         cloudformation.StackManager
	certificateDescriber iam.CertificateDescriber
	ec2Client            ec2.Client
	cloudFormationClient cloudformation.Client
	elbClient            *elb.ELB
}

func NewAWS(configuration acceptance.Config) AWS {
	awsConfig := aws.Config{
		AccessKeyID:     configuration.AWSAccessKeyID,
		SecretAccessKey: configuration.AWSSecretAccessKey,
		Region:          configuration.AWSRegion,
	}

	clientProvider := &clientmanager.ClientProvider{}
	clientProvider.SetConfig(awsConfig)

	stackManager := cloudformation.NewStackManager(clientProvider, application.NewLogger(os.Stdout))
	certificateDescriber := iam.NewCertificateDescriber(clientProvider)

	return AWS{
		stackManager:         stackManager,
		certificateDescriber: certificateDescriber,
		ec2Client:            clientProvider.GetEC2Client(),
		cloudFormationClient: clientProvider.GetCloudFormationClient(),
		elbClient:            elb.New(session.New(awsConfig.ClientConfig())),
	}
}

func (a AWS) StackExists(stackName string) bool {
	_, err := a.stackManager.Describe(stackName)

	if err == cloudformation.StackNotFound {
		return false
	}

	Expect(err).NotTo(HaveOccurred())
	return true
}

func (a AWS) Instances(envID string) []string {
	var instances []string

	vpcID := a.GetVPC(fmt.Sprintf("%s-vpc", envID))

	output, err := a.ec2Client.DescribeInstances(&awsec2.DescribeInstancesInput{
		Filters: []*awsec2.Filter{
			{
				Name: awslib.String("vpc-id"),
				Values: []*string{
					vpcID,
				},
			},
		},
	})
	Expect(err).NotTo(HaveOccurred())

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				if awslib.StringValue(tag.Key) == "Name" && awslib.StringValue(tag.Value) != "" {
					instances = append(instances, awslib.StringValue(tag.Value))
				}
			}
		}
	}

	return instances
}

func (a AWS) LoadBalancers(vpcName string) []string {
	var loadBalancerNames []string

	vpcID := a.GetVPC(vpcName)

	loadBalancerOutput, err := a.elbClient.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	Expect(err).NotTo(HaveOccurred())

	for _, lbDescription := range loadBalancerOutput.LoadBalancerDescriptions {
		if *lbDescription.VPCId == *vpcID {
			loadBalancerNames = append(loadBalancerNames, *lbDescription.LoadBalancerName)
		}
	}

	return loadBalancerNames
}

func (a AWS) DescribeCertificate(certificateName string) iam.Certificate {
	certificate, err := a.certificateDescriber.Describe(certificateName)
	if err != nil && err != iam.CertificateNotFound {
		Expect(err).NotTo(HaveOccurred())
	}

	return certificate
}

func (a AWS) GetEC2InstanceTags(instanceName string) map[string]string {
	output, err := a.ec2Client.DescribeInstances(&awsec2.DescribeInstancesInput{})
	Expect(err).NotTo(HaveOccurred())

	for _, reservation := range output.Reservations {
		for _, instance := range reservation.Instances {
			for _, tag := range instance.Tags {
				if awslib.StringValue(tag.Key) == "Name" && awslib.StringValue(tag.Value) == instanceName {
					tags := make(map[string]string)
					for _, tag := range instance.Tags {
						tags[awslib.StringValue(tag.Key)] = awslib.StringValue(tag.Value)
					}
					return tags
				}
			}
		}
	}
	return map[string]string{}
}

func (a AWS) GetSSLCertificateNameFromLBs(envID string) string {
	loadBalancerOutput, err := a.elbClient.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	Expect(err).NotTo(HaveOccurred())

	vpcID := a.GetVPC(fmt.Sprintf("%s-vpc", envID))

	var certificateName string
	for _, lbDescription := range loadBalancerOutput.LoadBalancerDescriptions {
		if *lbDescription.VPCId == *vpcID {
			for _, ld := range lbDescription.ListenerDescriptions {
				if int(*ld.Listener.LoadBalancerPort) == 443 {
					certificateArn := ld.Listener.SSLCertificateId
					certificateArnParts := strings.Split(awslib.StringValue(certificateArn), "/")
					certificateName = certificateArnParts[1]
					Expect(certificateName).NotTo(BeEmpty())

					return certificateName
				}
			}
		}
	}

	return ""
}

func (a AWS) GetVPC(vpcName string) *string {
	vpcs, err := a.ec2Client.DescribeVpcs(&awsec2.DescribeVpcsInput{
		Filters: []*awsec2.Filter{
			{
				Name: awslib.String("tag:Name"),
				Values: []*string{
					awslib.String(vpcName),
				},
			},
		},
	})

	Expect(err).NotTo(HaveOccurred())
	Expect(vpcs.Vpcs).To(HaveLen(1))

	return vpcs.Vpcs[0].VpcId
}

func (a AWS) DescribeKeyPairs(keypairName string) []*awsec2.KeyPairInfo {
	params := &awsec2.DescribeKeyPairsInput{
		Filters: []*awsec2.Filter{{}},
		KeyNames: []*string{
			awslib.String(keypairName),
		},
	}

	keypairOutput, err := a.ec2Client.DescribeKeyPairs(params)
	Expect(err).NotTo(HaveOccurred())

	return keypairOutput.KeyPairs
}

func (a AWS) NetworkHasBOSHDirector(envID string) bool {
	instances := a.Instances(envID)

	for _, instance := range instances {
		if instance == "bosh/0" {
			return true
		}
	}

	return false
}
