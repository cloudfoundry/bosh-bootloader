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

	// Usage Command
	usage := commands.NewUsage(os.Stdout)

	commandLineParser := application.NewCommandLineParser(usage.Print)
	configurationParser := application.NewConfigurationParser(commandLineParser, stateStore)
	configuration, err := configurationParser.Parse(os.Args[1:])
	if err != nil {
		fail(err)
	}

	// Amazon
	awsConfiguration := aws.Config{
		AccessKeyID:      configuration.State.AWS.AccessKeyID,
		SecretAccessKey:  configuration.State.AWS.SecretAccessKey,
		Region:           configuration.State.AWS.Region,
		EndpointOverride: configuration.Global.EndpointOverride,
	}

	cloudFormationClient := cloudformation.NewClient(awsConfiguration)
	ec2Client := ec2.NewClient(awsConfiguration)
	iamClient := iam.NewClient(awsConfiguration)

	awsCredentialValidator := application.NewAWSCredentialValidator(configuration)
	vpcStatusChecker := ec2.NewVPCStatusChecker(ec2Client)
	keyPairCreator := ec2.NewKeyPairCreator(ec2Client, uuidGenerator)
	keyPairDeleter := ec2.NewKeyPairDeleter(ec2Client, logger)
	keyPairChecker := ec2.NewKeyPairChecker(ec2Client)
	keyPairManager := ec2.NewKeyPairManager(keyPairCreator, keyPairChecker, logger)
	keyPairSynchronizer := ec2.NewKeyPairSynchronizer(keyPairManager)
	availabilityZoneRetriever := ec2.NewAvailabilityZoneRetriever(ec2Client)
	templateBuilder := templates.NewTemplateBuilder(logger)
	stackManager := cloudformation.NewStackManager(cloudFormationClient, logger)
	infrastructureManager := cloudformation.NewInfrastructureManager(templateBuilder, stackManager)
	certificateUploader := iam.NewCertificateUploader(iamClient, uuidGenerator, logger)
	certificateDescriber := iam.NewCertificateDescriber(iamClient)
	certificateDeleter := iam.NewCertificateDeleter(iamClient, logger)
	certificateManager := iam.NewCertificateManager(certificateUploader, certificateDescriber, certificateDeleter)
	certificateValidator := iam.NewCertificateValidator()

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
	boshClientProvider := bosh.NewClientProvider()
	cloudConfigGenerator := bosh.NewCloudConfigGenerator()
	cloudConfigurator := bosh.NewCloudConfigurator(logger, cloudConfigGenerator)
	cloudConfigManager := bosh.NewCloudConfigManager(logger, cloudConfigGenerator)

	// Commands
	help := commands.NewUsage(os.Stdout)
	version := commands.NewVersion(os.Stdout)
	up := commands.NewUp(
		awsCredentialValidator, infrastructureManager, keyPairSynchronizer, boshinitExecutor,
		stringGenerator, cloudConfigurator, availabilityZoneRetriever, certificateDescriber,
		cloudConfigManager, boshClientProvider,
	)
	destroy := commands.NewDestroy(
		awsCredentialValidator, logger, os.Stdin, boshinitExecutor, vpcStatusChecker, stackManager,
		stringGenerator, infrastructureManager, keyPairDeleter, certificateDeleter,
	)
	createLBs := commands.NewCreateLBs(
		logger, awsCredentialValidator, certificateManager, infrastructureManager,
		availabilityZoneRetriever, boshClientProvider, cloudConfigurator, cloudConfigManager, certificateValidator,
	)
	updateLBs := commands.NewUpdateLBs(awsCredentialValidator, certificateManager,
		availabilityZoneRetriever, infrastructureManager, boshClientProvider, logger,
	)
	deleteLBs := commands.NewDeleteLBs(
		awsCredentialValidator, availabilityZoneRetriever, certificateManager,
		infrastructureManager, logger, cloudConfigurator, cloudConfigManager, boshClientProvider,
	)
	lbs := commands.NewLBs(awsCredentialValidator, infrastructureManager, os.Stdout)
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
		"destroy":                   destroy,
		"director-address":          directorAddress,
		"director-username":         directorUsername,
		"director-password":         directorPassword,
		"ssh-key":                   sshKey,
		commands.CREATE_LBS_COMMAND: createLBs,
		"unsupported-update-lbs":    updateLBs,
		"unsupported-delete-lbs":    deleteLBs,
		"lbs": lbs,
	}, configuration, stateStore, usage.Print)

	err = app.Run()
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "\n\n%s\n", err)
	os.Exit(1)
}
