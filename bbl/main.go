package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation/templates"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/ec2"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/elb"
	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/iam"
	"github.com/pivotal-cf-experimental/bosh-bootloader/bosh"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit"
	"github.com/pivotal-cf-experimental/bosh-bootloader/boshinit/manifests"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/helpers"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ssl"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

func main() {
	// Utilities
	uuidGenerator := helpers.NewUUIDGenerator(rand.Reader)
	stringGenerator := helpers.NewStringGenerator(rand.Reader)
	logger := application.NewLogger(os.Stdout)
	stateStore := storage.NewStore()
	sslKeyPairGenerator := ssl.NewKeyPairGenerator(time.Now, rsa.GenerateKey, x509.CreateCertificate)

	// Amazon
	awsClientProvider := aws.NewClientProvider()
	vpcStatusChecker := ec2.NewVPCStatusChecker()
	keyPairCreator := ec2.NewKeyPairCreator(uuidGenerator)
	keyPairDeleter := ec2.NewKeyPairDeleter(logger)
	keyPairChecker := ec2.NewKeyPairChecker()
	keyPairManager := ec2.NewKeyPairManager(keyPairCreator, keyPairChecker, logger)
	keyPairSynchronizer := ec2.NewKeyPairSynchronizer(keyPairManager)
	templateBuilder := templates.NewTemplateBuilder(logger)
	stackManager := cloudformation.NewStackManager(logger)
	infrastructureManager := cloudformation.NewInfrastructureManager(templateBuilder, stackManager)
	elbDescriber := elb.NewDescriber()
	certificateUploader := iam.NewCertificateUploader()
	certificateDescriber := iam.NewCertificateDescriber()
	certificateDeleter := iam.NewCertificateDeleter()
	certificateManager := iam.NewCertificateManager(certificateUploader, certificateDescriber, certificateDeleter)

	// bosh-init
	tempDir, err := ioutil.TempDir("", "bosh-init")
	if err != nil {
		fail(err)
	}

	boshInitPath, err := exec.LookPath("bosh-init")
	if err != nil {
		fail(err)
	}

	cloudProviderManifestBuilder := manifests.NewCloudProviderManifestBuilder(stringGenerator)
	jobsManifestBuilder := manifests.NewJobsManifestBuilder(stringGenerator)
	boshinitManifestBuilder := manifests.NewManifestBuilder(
		logger, sslKeyPairGenerator, stringGenerator, cloudProviderManifestBuilder, jobsManifestBuilder,
	)
	boshinitCommandBuilder := boshinit.NewCommandBuilder(boshInitPath, tempDir, os.Stdout, os.Stderr)
	boshinitDeployCommand := boshinitCommandBuilder.DeployCommand()
	boshinitDeleteCommand := boshinitCommandBuilder.DeleteCommand()
	boshinitDeployRunner := boshinit.NewCommandRunner(tempDir, boshinitDeployCommand)
	boshinitDeleteRunner := boshinit.NewCommandRunner(tempDir, boshinitDeleteCommand)
	boshinitExecutor := boshinit.NewExecutor(
		boshinitManifestBuilder, boshinitDeployRunner, boshinitDeleteRunner, logger,
	)

	// BOSH
	boshCloudConfigGenerator := bosh.NewCloudConfigGenerator()
	cloudConfigurator := bosh.NewCloudConfigurator(logger, boshCloudConfigGenerator)
	availabilityZoneRetriever := ec2.NewAvailabilityZoneRetriever()

	// Commands
	help := commands.NewUsage(os.Stdout)
	version := commands.NewVersion(os.Stdout)
	up := commands.NewUp(
		infrastructureManager, keyPairSynchronizer, awsClientProvider, boshinitExecutor,
		stringGenerator, cloudConfigurator, availabilityZoneRetriever, elbDescriber, certificateManager,
	)
	destroy := commands.NewDestroy(
		logger, os.Stdin, boshinitExecutor, awsClientProvider, vpcStatusChecker,
		stackManager, stringGenerator, infrastructureManager, keyPairDeleter,
	)
	usage := commands.NewUsage(os.Stdout)
	directorAddress := commands.NewStateQuery(logger, "director address", func(state storage.State) string {
		return state.BOSH.DirectorAddress
	})
	directorUsername := commands.NewStateQuery(logger, "director username", func(state storage.State) string {
		return state.BOSH.DirectorUsername
	})
	directorPassword := commands.NewStateQuery(logger, "director password", func(state storage.State) string {
		return state.BOSH.DirectorPassword
	})
	sshKey := commands.NewStateQuery(logger, "ssh key", func(state storage.State) string {
		return state.KeyPair.PrivateKey
	})

	app := application.New(application.CommandSet{
		"help":    help,
		"version": version,
		"unsupported-deploy-bosh-on-aws-for-concourse": up,
		"destroy":           destroy,
		"director-address":  directorAddress,
		"director-username": directorUsername,
		"director-password": directorPassword,
		"ssh-key":           sshKey,
	}, stateStore, usage.Print)

	err = app.Run(os.Args[1:])
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "\n\n%s\n", err)
	os.Exit(1)
}
