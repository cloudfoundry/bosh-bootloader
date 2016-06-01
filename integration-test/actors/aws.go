package actors

import (
	"os"

	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
)

type AWS struct {
	stackManager         cloudformation.StackManager
	certificateDescriber iam.CertificateDescriber
}

func NewAWS(configuration integration.Config) AWS {
	cloudFormationClient := cloudformation.NewClient(aws.Config{
		AccessKeyID:     configuration.AWSAccessKeyID,
		SecretAccessKey: configuration.AWSSecretAccessKey,
		Region:          configuration.AWSRegion,
	})

	iamClient := iam.NewClient(aws.Config{
		AccessKeyID:     configuration.AWSAccessKeyID,
		SecretAccessKey: configuration.AWSSecretAccessKey,
		Region:          configuration.AWSRegion,
	})

	stackManager := cloudformation.NewStackManager(cloudFormationClient, application.NewLogger(os.Stdout))
	certificateDescriber := iam.NewCertificateDescriber(iamClient)

	return AWS{
		stackManager:         stackManager,
		certificateDescriber: certificateDescriber,
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

func (a AWS) LoadBalancers(stackName string) map[string]string {
	stack, err := a.stackManager.Describe(stackName)
	Expect(err).NotTo(HaveOccurred())

	loadBalancers := map[string]string{}

	for _, loadBalancer := range []string{"CFRouterLoadBalancer", "CFSSHProxyLoadBalancer", "ConcourseLoadBalancer"} {
		if stack.Outputs[loadBalancer] != "" {
			loadBalancers[loadBalancer] = stack.Outputs[loadBalancer]
		}
	}

	return loadBalancers
}

func (a AWS) DescribeCertificate(certificateName string) iam.Certificate {
	certificate, err := a.certificateDescriber.Describe(certificateName)
	if err != nil && err != iam.CertificateNotFound {
		Expect(err).NotTo(HaveOccurred())
	}

	return certificate
}
