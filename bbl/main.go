package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/clientmanager"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/boshinit"
	"github.com/cloudfoundry/bosh-bootloader/boshinit/manifests"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/ssl"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/square/certstrap/pkix"
)

func main() {
	// Utilities
	uuidGenerator := helpers.NewUUIDGenerator(rand.Reader)
	stringGenerator := helpers.NewStringGenerator(rand.Reader)
	envIDGenerator := helpers.NewEnvIDGenerator(rand.Reader)
	logger := application.NewLogger(os.Stdout)
	sslKeyPairGenerator := ssl.NewKeyPairGenerator(rsa.GenerateKey, pkix.CreateCertificateAuthority, pkix.CreateCertificateSigningRequest, pkix.CreateCertificateHost)

	// Usage Command
	usage := commands.NewUsage(os.Stdout)

	commandLineParser := application.NewCommandLineParser(usage.Print)
	configurationParser := application.NewConfigurationParser(commandLineParser)
	configuration, err := configurationParser.Parse(os.Args[1:])
	if err != nil {
		fail(err)
	}

	stateStore := storage.NewStore(configuration.Global.StateDir)

	// Amazon
	awsConfiguration := aws.Config{
		AccessKeyID:      configuration.State.AWS.AccessKeyID,
		SecretAccessKey:  configuration.State.AWS.SecretAccessKey,
		Region:           configuration.State.AWS.Region,
		EndpointOverride: configuration.Global.EndpointOverride,
	}

	clientProvider := &clientmanager.ClientProvider{}
	clientProvider.SetConfig(awsConfiguration)

	awsCredentialValidator := application.NewAWSCredentialValidator(configuration)
	vpcStatusChecker := ec2.NewVPCStatusChecker(clientProvider)
	keyPairCreator := ec2.NewKeyPairCreator(clientProvider)
	keyPairDeleter := ec2.NewKeyPairDeleter(clientProvider, logger)
	keyPairChecker := ec2.NewKeyPairChecker(clientProvider)
	keyPairManager := ec2.NewKeyPairManager(keyPairCreator, keyPairChecker, logger)
	keyPairSynchronizer := ec2.NewKeyPairSynchronizer(keyPairManager)
	availabilityZoneRetriever := ec2.NewAvailabilityZoneRetriever(clientProvider)
	templateBuilder := templates.NewTemplateBuilder(logger)
	stackManager := cloudformation.NewStackManager(clientProvider, logger)
	infrastructureManager := cloudformation.NewInfrastructureManager(templateBuilder, stackManager)
	certificateUploader := iam.NewCertificateUploader(clientProvider)
	certificateDescriber := iam.NewCertificateDescriber(clientProvider)
	certificateDeleter := iam.NewCertificateDeleter(clientProvider)
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
		cloudConfigManager, boshClientProvider, envIDGenerator, stateStore,
		clientProvider,
	)
	destroy := commands.NewDestroy(
		awsCredentialValidator, logger, os.Stdin, boshinitExecutor, vpcStatusChecker, stackManager,
		stringGenerator, infrastructureManager, keyPairDeleter, certificateDeleter, stateStore,
	)
	createLBs := commands.NewCreateLBs(
		logger, awsCredentialValidator, certificateManager, infrastructureManager,
		availabilityZoneRetriever, boshClientProvider, cloudConfigurator, cloudConfigManager, certificateValidator,
		uuidGenerator, stateStore,
	)
	updateLBs := commands.NewUpdateLBs(awsCredentialValidator, certificateManager,
		availabilityZoneRetriever, infrastructureManager, boshClientProvider, logger, certificateValidator, uuidGenerator,
		stateStore)
	deleteLBs := commands.NewDeleteLBs(
		awsCredentialValidator, availabilityZoneRetriever, certificateManager,
		infrastructureManager, logger, cloudConfigurator, cloudConfigManager, boshClientProvider, stateStore,
	)
	lbs := commands.NewLBs(awsCredentialValidator, infrastructureManager, os.Stdout)
	directorAddress := commands.NewStateQuery(logger, commands.DirectorAddressPropertyName, func(state storage.State) string {
		return state.BOSH.DirectorAddress
	})
	directorUsername := commands.NewStateQuery(logger, commands.DirectorUsernamePropertyName, func(state storage.State) string {
		return state.BOSH.DirectorUsername
	})
	directorPassword := commands.NewStateQuery(logger, commands.DirectorPasswordPropertyName, func(state storage.State) string {
		return state.BOSH.DirectorPassword
	})
	sshKey := commands.NewStateQuery(logger, commands.SSHKeyPropertyName, func(state storage.State) string {
		return state.KeyPair.PrivateKey
	})
	boshCACert := commands.NewStateQuery(logger, commands.BOSHCACertPropertyName, func(state storage.State) string {
		return state.BOSH.DirectorSSLCA
	})
	envID := commands.NewStateQuery(logger, commands.EnvIDPropertyName, func(state storage.State) string {
		return state.EnvID
	})

	app := application.New(application.CommandSet{
		commands.HelpCommand:             help,
		commands.VersionCommand:          version,
		commands.UpCommand:               up,
		commands.DestroyCommand:          destroy,
		commands.DirectorAddressCommand:  directorAddress,
		commands.DirectorUsernameCommand: directorUsername,
		commands.DirectorPasswordCommand: directorPassword,
		commands.SSHKeyCommand:           sshKey,
		commands.CreateLBsCommand:        createLBs,
		commands.UpdateLBsCommand:        updateLBs,
		commands.DeleteLBsCommand:        deleteLBs,
		commands.LBsCommand:              lbs,
		commands.BOSHCACertCommand:       boshCACert,
		commands.EnvIDCommand:            envID,
	}, configuration, stateStore, usage)

	err = app.Run()
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "\n\n%s\n", err)
	os.Exit(1)
}
