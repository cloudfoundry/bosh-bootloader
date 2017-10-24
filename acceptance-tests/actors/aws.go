package actors

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"

	. "github.com/onsi/gomega"

	awslib "github.com/aws/aws-sdk-go/aws"
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
)

type AWS struct {
	client    ec2.Client
	elbClient *elb.ELB
}

func NewAWS(configuration acceptance.Config) AWS {
	awsConfig := aws.Config{
		AccessKeyID:     configuration.AWSAccessKeyID,
		SecretAccessKey: configuration.AWSSecretAccessKey,
		Region:          configuration.AWSRegion,
	}

	client := ec2.NewClient(awsConfig, application.NewLogger(os.Stdout))

	return AWS{
		client:    client,
		elbClient: elb.New(session.New(awsConfig.ClientConfig())),
	}
}

func (a AWS) Instances(envID string) []string {
	instances, err := a.client.Instances(envID)
	Expect(err).NotTo(HaveOccurred())
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

func (a AWS) GetSSLCertificateNameFromLBs(envID string) string {
	var retryCount int

retry:

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
					if len(certificateArnParts) != 2 && retryCount <= 5 {
						retryCount++
						goto retry
					}
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
	vpc, err := a.client.GetVPC(vpcName)
	Expect(err).NotTo(HaveOccurred())
	return vpc
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
