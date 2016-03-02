package main

import (
	"crypto/rand"
	"fmt"
	"os"

	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

func main() {
	uuidGenerator := ec2.NewUUIDGenerator(rand.Reader)
	logger := application.NewLogger(os.Stdout)

	templateBuilder := cloudformation.NewTemplateBuilder(logger)
	keypairCreator := ec2.NewKeyPairCreator(uuidGenerator.Generate)
	keypairRetriever := ec2.NewKeyPairRetriever()
	keypairManager := ec2.NewKeyPairManager(keypairCreator, keypairRetriever, logger)
	stateStore := storage.NewStore()
	stackManager := cloudformation.NewStackManager(logger)
	awsClientProvider := aws.NewClientProvider()

	app := application.New(application.CommandSet{
		"help":    commands.NewUsage(os.Stdout),
		"version": commands.NewVersion(os.Stdout),
		"unsupported-provision-aws-for-concourse": unsupported.NewProvisionAWSForConcourse(templateBuilder, stackManager, keypairManager, awsClientProvider),
	}, stateStore, commands.NewUsage(os.Stdout).Print)

	err := app.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\n%s\n", err)
		os.Exit(1)
	}
}
