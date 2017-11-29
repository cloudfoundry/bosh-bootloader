package actors

import (
	"fmt"
	"os"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/storage"

	. "github.com/onsi/gomega"

	awslib "github.com/aws/aws-sdk-go/aws"
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/aws/aws-sdk-go/service/elbv2"
)

type AWS struct {
	client      aws.Client
	elbClient   *elb.ELB
	elbV2Client *elbv2.ELBV2
}

func NewAWS(c acceptance.Config) AWS {
	creds := storage.AWS{
		AccessKeyID:     c.AWSAccessKeyID,
		SecretAccessKey: c.AWSSecretAccessKey,
		Region:          c.AWSRegion,
	}
	client := aws.NewClient(creds, application.NewLogger(os.Stdout))

	elbConfig := &awslib.Config{
		Credentials: credentials.NewStaticCredentials(creds.AccessKeyID, creds.SecretAccessKey, ""),
		Region:      awslib.String(creds.Region),
	}
	return AWS{
		client:      client,
		elbClient:   elb.New(session.New(elbConfig)),
		elbV2Client: elbv2.New(session.New(elbConfig)),
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

func (a AWS) NetworkLoadBalancers(vpcName string) []string {
	output, err := a.elbV2Client.DescribeLoadBalancers(&elbv2.DescribeLoadBalancersInput{})
	Expect(err).NotTo(HaveOccurred())

	vpcId := a.GetVPC(vpcName)
	lbNames := []string{}
	for _, lb := range output.LoadBalancers {
		if *lb.VpcId == *vpcId {
			lbNames = append(lbNames, *lb.LoadBalancerName)
		}
	}

	return lbNames
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
