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

func NewAWSLBHelper(c acceptance.Config) awsLbHelper {
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
	return awsLbHelper{
		client:    client,
		elbClient: elb.New(session.New(elbConfig)),
	}
}

func (a awsLbHelper) loadBalancers(vpcName string) []string {
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

type awsLbHelper struct {
	client    aws.Client
	elbClient *elb.ELB
}

func (a awsLbHelper) GetLBArgs() []string {
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

func (a awsLbHelper) ConfirmLBsExist(envID string) {
	vpcName := fmt.Sprintf("%s-vpc", envID)
	Expect(a.loadBalancers(vpcName)).To(HaveLen(3))
	Expect(a.loadBalancers(vpcName)).To(ConsistOf(
		MatchRegexp(".*-cf-router-lb"),
		MatchRegexp(".*-cf-ssh-lb"),
		MatchRegexp(".*-cf-tcp-lb"),
	))
}

func (a awsLbHelper) ConfirmNoLBsExist(envID string) {
	vpcName := fmt.Sprintf("%s-vpc", envID)
	Expect(a.loadBalancers(vpcName)).To(BeEmpty())
}

func (a awsLbHelper) VerifyBblLBOutput(stdout string) {
	Expect(stdout).To(MatchRegexp("CF Router LB:.*"))
	Expect(stdout).To(MatchRegexp("CF SSH Proxy LB:.*"))
	Expect(stdout).To(MatchRegexp("CF TCP Router LB:.*"))
}
