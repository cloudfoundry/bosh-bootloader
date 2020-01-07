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

func NewAWSLBHelper(c acceptance.Config) awsLBHelper {
	creds := storage.AWS{
		AccessKeyID:     c.AWSAccessKeyID,
		SecretAccessKey: c.AWSSecretAccessKey,
		Region:          c.AWSRegion,
	}

	logger := application.NewLogger(os.Stdout, os.Stdin)

	client := aws.NewClient(creds, logger)

	elbConfig := &awslib.Config{
		Credentials: credentials.NewStaticCredentials(creds.AccessKeyID, creds.SecretAccessKey, ""),
		Region:      awslib.String(creds.Region),
	}
	return awsLBHelper{
		client:    client,
		elbClient: elb.New(session.New(elbConfig)),
	}
}

func (a awsLBHelper) loadBalancers(vpcName string) []string {
	var loadBalancerNames []string

	vpcID, err := a.client.GetVPC(vpcName)
	Expect(err).NotTo(HaveOccurred())

	loadBalancerOutput, err := a.elbClient.DescribeLoadBalancers(&elb.DescribeLoadBalancersInput{})
	Expect(err).NotTo(HaveOccurred())

	for _, lbDescription := range loadBalancerOutput.LoadBalancerDescriptions {
		if *lbDescription.VPCId == *vpcID {
			loadBalancerNames = append(loadBalancerNames, *lbDescription.LoadBalancerName)
		}
	}

	return loadBalancerNames
}

type awsLBHelper struct {
	client    aws.Client
	elbClient *elb.ELB
}

func (a awsLBHelper) GetLBArgs() []string {
	certPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_CERT)
	Expect(err).NotTo(HaveOccurred())
	keyPath, err := testhelpers.WriteContentsToTempFile(testhelpers.BBL_KEY)
	Expect(err).NotTo(HaveOccurred())

	return []string{
		"--lb-type", "cf",
		"--lb-cert", certPath,
		"--lb-key", keyPath,
	}
}

func (a awsLBHelper) VerifyCloudConfigExtensions(vmExtensions []string) {
	Expect(vmExtensions).To(ContainElement("cf-router-network-properties"))
	Expect(vmExtensions).To(ContainElement("diego-ssh-proxy-network-properties"))
	Expect(vmExtensions).To(ContainElement("cf-tcp-router-network-properties"))
}

func (a awsLBHelper) ConfirmLBsExist(envID string) {
	vpcName := fmt.Sprintf("%s-vpc", envID)
	Expect(a.loadBalancers(vpcName)).To(HaveLen(3))
	Expect(a.loadBalancers(vpcName)).To(ConsistOf(
		MatchRegexp(".*-cf-router-lb"),
		MatchRegexp(".*-cf-ssh-lb"),
		MatchRegexp(".*-cf-tcp-lb"),
	))
}

func (a awsLBHelper) ConfirmNoLBsExist(envID string) {
	vpcName := fmt.Sprintf("%s-vpc", envID)
	Expect(a.loadBalancers(vpcName)).To(BeEmpty())
}

func (a awsLBHelper) VerifyBblLBOutput(stdout string) {
	Expect(stdout).To(MatchRegexp("CF Router LB:.*"))
	Expect(stdout).To(MatchRegexp("CF SSH Proxy LB:.*"))
	Expect(stdout).To(MatchRegexp("CF TCP Router LB:.*"))
}

func (a awsLBHelper) ConfirmNoStemcellsExist(stemcellIDs []string) {}
