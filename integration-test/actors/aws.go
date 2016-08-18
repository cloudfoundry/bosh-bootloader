package actors

import (
	"os"

	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/integration-test"

	. "github.com/onsi/gomega"

	awslib "github.com/aws/aws-sdk-go/aws"
	awsec2 "github.com/aws/aws-sdk-go/service/ec2"
)

type AWS struct {
	stackManager         cloudformation.StackManager
	certificateDescriber iam.CertificateDescriber
	ec2Client            ec2.Client
	cloudFormationClient cloudformation.Client
}

func NewAWS(configuration integration.Config) AWS {
	awsConfig := aws.Config{
		AccessKeyID:     configuration.AWSAccessKeyID,
		SecretAccessKey: configuration.AWSSecretAccessKey,
		Region:          configuration.AWSRegion,
	}

	iamClient := iam.NewClient(awsConfig)
	ec2Client := ec2.NewClient(awsConfig)
	cloudFormationClient := cloudformation.NewClient(awsConfig)

	stackManager := cloudformation.NewStackManager(cloudFormationClient, application.NewLogger(os.Stdout))
	certificateDescriber := iam.NewCertificateDescriber(iamClient)

	return AWS{
		stackManager:         stackManager,
		certificateDescriber: certificateDescriber,
		ec2Client:            ec2Client,
		cloudFormationClient: cloudFormationClient,
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

func (a AWS) GetPhysicalID(stackName, logicalID string) string {
	physicalID, err := a.stackManager.GetPhysicalIDForResource(stackName, logicalID)
	Expect(err).NotTo(HaveOccurred())
	return physicalID
}

func (a AWS) LoadBalancers(stackName string) map[string]string {
	stack, err := a.stackManager.Describe(stackName)
	Expect(err).NotTo(HaveOccurred())

	loadBalancers := map[string]string{}

	for _, loadBalancer := range []string{"CFRouterLoadBalancer", "CFSSHProxyLoadBalancer", "ConcourseLoadBalancer", "ConcourseLoadBalancerURL"} {
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

func (a AWS) GetEC2InstanceTags(instanceID string) map[string]string {
	describeInstanceInput := &awsec2.DescribeInstancesInput{
		DryRun: awslib.Bool(false),
		Filters: []*awsec2.Filter{
			{
				Name: awslib.String("instance-id"),
				Values: []*string{
					awslib.String(instanceID),
				},
			},
		},
		InstanceIds: []*string{
			awslib.String(instanceID),
		},
	}
	describeInstancesOutput, err := a.ec2Client.DescribeInstances(describeInstanceInput)
	Expect(err).NotTo(HaveOccurred())
	Expect(describeInstancesOutput.Reservations).To(HaveLen(1))
	Expect(describeInstancesOutput.Reservations[0].Instances).To(HaveLen(1))

	instance := describeInstancesOutput.Reservations[0].Instances[0]

	tags := make(map[string]string)
	for _, tag := range instance.Tags {
		tags[awslib.StringValue(tag.Key)] = awslib.StringValue(tag.Value)
	}
	return tags
}
