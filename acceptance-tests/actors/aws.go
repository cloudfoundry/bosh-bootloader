package actors

import (
	"fmt"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/testhelpers"

	. "github.com/onsi/gomega"

	awslib "github.com/aws/aws-sdk-go/aws"
	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/elb"
)

type AWS struct {
	client    aws.Client
	elbClient *elb.ELB
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
		client:    client,
		elbClient: elb.New(session.New(elbConfig)),
	}
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

func (a AWS) GetVPC(vpcName string) *string {
	vpc, err := a.client.GetVPC(vpcName)
	Expect(err).NotTo(HaveOccurred())
	return vpc
}

type awsIaasLbHelper struct {
	aws AWS
}

func (a awsIaasLbHelper) GetLBArgs() []string {
	certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
	Expect(err).NotTo(HaveOccurred())
	chainPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CHAIN)
	Expect(err).NotTo(HaveOccurred())
	keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
	Expect(err).NotTo(HaveOccurred())

	return []string{
		"--lb-type", "cf",
		"--lb-cert", certPath,
		"--lb-key", keyPath,
		"--lb-chain", chainPath,
	}
}

func (a awsIaasLbHelper) ConfirmLBsExist(envID string) {
	vpcName := fmt.Sprintf("%s-vpc", envID)
	Expect(a.aws.LoadBalancers(vpcName)).To(HaveLen(3))
	Expect(a.aws.LoadBalancers(vpcName)).To(ConsistOf(
		MatchRegexp(".*-cf-router-lb"),
		MatchRegexp(".*-cf-ssh-lb"),
		MatchRegexp(".*-cf-tcp-lb"),
	))
}

func (a awsIaasLbHelper) ConfirmNoLBsExist(envID string) {
	vpcName := fmt.Sprintf("%s-vpc", envID)
	Expect(a.aws.LoadBalancers(vpcName)).To(BeEmpty())
}
