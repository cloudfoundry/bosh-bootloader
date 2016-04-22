package actors

import (
	"os"

	. "github.com/onsi/gomega"
	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"
)

type AWS struct {
	stackManager         cloudformation.StackManager
	cloudFormationClient cloudformation.Client
}

func NewAWS(configuration integration.Config) AWS {
	stackManager := cloudformation.NewStackManager(application.NewLogger(os.Stdout))
	cloudFormationClient, err := aws.NewClientProvider().CloudFormationClient(aws.Config{
		AccessKeyID:     configuration.AWSAccessKeyID,
		SecretAccessKey: configuration.AWSSecretAccessKey,
		Region:          configuration.AWSRegion,
	})
	Expect(err).NotTo(HaveOccurred())

	return AWS{
		stackManager:         stackManager,
		cloudFormationClient: cloudFormationClient,
	}
}

func (a AWS) StackExists(stackName string) bool {
	_, err := a.stackManager.Describe(a.cloudFormationClient, stackName)

	if err == cloudformation.StackNotFound {
		return false
	}

	Expect(err).NotTo(HaveOccurred())
	return true
}

func (a AWS) LoadBalancers(stackName string) []string {
	stack, err := a.stackManager.Describe(a.cloudFormationClient, stackName)
	Expect(err).NotTo(HaveOccurred())

	loadBalancers := []string{}

	for _, loadBalancer := range []string{"CFRouterLoadBalancer", "CFSSHProxyLoadBalancer", "ConcourseLoadBalancer"} {
		if stack.Outputs[loadBalancer] != "" {
			loadBalancers = append(loadBalancers, loadBalancer)
		}
	}

	return loadBalancers
}
