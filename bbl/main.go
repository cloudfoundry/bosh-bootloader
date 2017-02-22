package main

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	yaml "gopkg.in/yaml.v2"

	"golang.org/x/crypto/ssh"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/clientmanager"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	gcpcloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
)

var (
	Version     string
	gcpBasePath string
)

func main() {
	// Command Set
	commandSet := application.CommandSet{
		commands.HelpCommand:             nil,
		commands.VersionCommand:          nil,
		commands.UpCommand:               nil,
		commands.DestroyCommand:          nil,
		commands.DirectorAddressCommand:  nil,
		commands.DirectorUsernameCommand: nil,
		commands.DirectorPasswordCommand: nil,
		commands.DirectorCACertCommand:   nil,
		commands.BOSHCACertCommand:       nil,
		commands.SSHKeyCommand:           nil,
		commands.CreateLBsCommand:        nil,
		commands.UpdateLBsCommand:        nil,
		commands.DeleteLBsCommand:        nil,
		commands.LBsCommand:              nil,
		commands.EnvIDCommand:            nil,
		commands.PrintEnvCommand:         nil,
	}

	// Utilities
	uuidGenerator := helpers.NewUUIDGenerator(rand.Reader)
	stringGenerator := helpers.NewStringGenerator(rand.Reader)
	envIDGenerator := helpers.NewEnvIDGenerator(rand.Reader)
	logger := application.NewLogger(os.Stdout)
	stderrLogger := application.NewLogger(os.Stderr)

	// Usage Command
	usage := commands.NewUsage(os.Stdout)
	storage.GetStateLogger = stderrLogger

	commandLineParser := application.NewCommandLineParser(usage.Print, commandSet)
	configurationParser := application.NewConfigurationParser(commandLineParser)
	configuration, err := configurationParser.Parse(os.Args[1:])
	if err != nil {
		fail(err)
	}

	stateStore := storage.NewStore(configuration.Global.StateDir)
	stateValidator := application.NewStateValidator(configuration.Global.StateDir)

	// Amazon
	awsConfiguration := aws.Config{
		AccessKeyID:      configuration.State.AWS.AccessKeyID,
		SecretAccessKey:  configuration.State.AWS.SecretAccessKey,
		Region:           configuration.State.AWS.Region,
		EndpointOverride: configuration.Global.EndpointOverride,
	}

	clientProvider := &clientmanager.ClientProvider{EndpointOverride: configuration.Global.EndpointOverride}
	clientProvider.SetConfig(awsConfiguration)

	credentialValidator := application.NewCredentialValidator(configuration)
	vpcStatusChecker := ec2.NewVPCStatusChecker(clientProvider)
	awsKeyPairCreator := ec2.NewKeyPairCreator(clientProvider)
	awsKeyPairDeleter := ec2.NewKeyPairDeleter(clientProvider, logger)
	keyPairChecker := ec2.NewKeyPairChecker(clientProvider)
	keyPairManager := ec2.NewKeyPairManager(awsKeyPairCreator, keyPairChecker, logger)
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

	// GCP
	gcpClientProvider := gcp.NewClientProvider(gcpBasePath)
	gcpClientProvider.SetConfig(configuration.State.GCP.ServiceAccountKey, configuration.State.GCP.ProjectID, configuration.State.GCP.Zone)

	gcpKeyPairUpdater := gcp.NewKeyPairUpdater(rand.Reader, rsa.GenerateKey, ssh.NewPublicKey, gcpClientProvider, logger)
	gcpCloudConfigGenerator := gcpcloudconfig.NewCloudConfigGenerator()
	gcpKeyPairDeleter := gcp.NewKeyPairDeleter(gcpClientProvider, logger)
	gcpNetworkInstancesChecker := gcp.NewNetworkInstancesChecker(gcpClientProvider)
	zones := gcp.NewZones()

	// EnvID
	envIDManager := helpers.NewEnvIDManager(envIDGenerator, gcpClientProvider, infrastructureManager)

	// Terraform
	terraformCmd := terraform.NewCmd(os.Stderr)
	terraformExecutor := terraform.NewExecutor(terraformCmd, configuration.Global.Debug)
	terraformOutputter := terraform.NewOutputter(terraformCmd)
	terraformOutputProvider := terraform.NewOutputProvider(terraformOutputter)

	// BOSH
	boshCommand := bosh.NewCmd(os.Stderr)
	boshExecutor := bosh.NewExecutor(boshCommand, ioutil.TempDir, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal,
		json.Marshal, ioutil.WriteFile, configuration.Global.Debug)
	boshManager := bosh.NewManager(boshExecutor, terraformOutputProvider, stackManager)
	boshClientProvider := bosh.NewClientProvider()
	cloudConfigGenerator := bosh.NewCloudConfigGenerator()
	cloudConfigurator := bosh.NewCloudConfigurator(logger, cloudConfigGenerator)
	cloudConfigManager := bosh.NewCloudConfigManager(logger, cloudConfigGenerator)

	// Subcommands
	awsUp := commands.NewAWSUp(
		credentialValidator, infrastructureManager, keyPairSynchronizer, boshManager,
		cloudConfigurator, availabilityZoneRetriever, certificateDescriber,
		cloudConfigManager, boshClientProvider, stateStore, clientProvider, envIDManager)

	awsCreateLBs := commands.NewAWSCreateLBs(
		logger, credentialValidator, certificateManager, infrastructureManager,
		availabilityZoneRetriever, boshClientProvider, cloudConfigurator, cloudConfigManager, certificateValidator,
		uuidGenerator, stateStore,
	)

	gcpCreateLBs := commands.NewGCPCreateLBs(terraformExecutor, terraformOutputProvider, gcpCloudConfigGenerator, boshClientProvider, zones, stateStore, logger)

	awsUpdateLBs := commands.NewAWSUpdateLBs(credentialValidator, certificateManager, availabilityZoneRetriever, infrastructureManager,
		boshClientProvider, logger, uuidGenerator, stateStore)

	gcpUpdateLBs := commands.NewGCPUpdateLBs(gcpCreateLBs)

	awsDeleteLBs := commands.NewAWSDeleteLBs(
		credentialValidator, availabilityZoneRetriever, certificateManager,
		infrastructureManager, logger, cloudConfigurator, cloudConfigManager, boshClientProvider, stateStore,
	)
	gcpDeleteLBs := commands.NewGCPDeleteLBs(terraformOutputProvider, gcpCloudConfigGenerator, zones, logger,
		boshClientProvider, stateStore, terraformExecutor)

	gcpUp := commands.NewGCPUp(stateStore, gcpKeyPairUpdater, gcpClientProvider, terraformExecutor, boshManager, logger, boshClientProvider,
		gcpCloudConfigGenerator, terraformOutputProvider, zones, envIDManager)
	envGetter := commands.NewEnvGetter()

	// Commands
	commandSet[commands.HelpCommand] = commands.NewUsage(os.Stdout)
	commandSet[commands.VersionCommand] = commands.NewVersion(Version, os.Stdout)

	commandSet[commands.UpCommand] = commands.NewUp(awsUp, gcpUp, envGetter)

	commandSet[commands.DestroyCommand] = commands.NewDestroy(
		credentialValidator, logger, os.Stdin, boshManager, vpcStatusChecker, stackManager,
		stringGenerator, infrastructureManager, awsKeyPairDeleter, gcpKeyPairDeleter, certificateDeleter,
		stateStore, stateValidator, terraformExecutor, terraformOutputProvider, gcpNetworkInstancesChecker,
	)

	commandSet[commands.CreateLBsCommand] = commands.NewCreateLBs(awsCreateLBs, gcpCreateLBs, stateValidator)
	commandSet[commands.UpdateLBsCommand] = commands.NewUpdateLBs(awsUpdateLBs, gcpUpdateLBs, certificateValidator, stateValidator, logger)
	commandSet[commands.DeleteLBsCommand] = commands.NewDeleteLBs(gcpDeleteLBs, awsDeleteLBs, logger, stateValidator)
	commandSet[commands.LBsCommand] = commands.NewLBs(credentialValidator, stateValidator, infrastructureManager, terraformOutputProvider, os.Stdout)
	commandSet[commands.DirectorAddressCommand] = commands.NewStateQuery(logger, stateValidator, commands.DirectorAddressPropertyName, func(state storage.State) string {
		return state.BOSH.DirectorAddress
	})
	commandSet[commands.DirectorUsernameCommand] = commands.NewStateQuery(logger, stateValidator, commands.DirectorUsernamePropertyName, func(state storage.State) string {
		return state.BOSH.DirectorUsername
	})
	commandSet[commands.DirectorPasswordCommand] = commands.NewStateQuery(logger, stateValidator, commands.DirectorPasswordPropertyName, func(state storage.State) string {
		return state.BOSH.DirectorPassword
	})
	commandSet[commands.DirectorCACertCommand] = commands.NewStateQuery(logger, stateValidator, commands.DirectorCACertPropertyName, func(state storage.State) string {
		return state.BOSH.DirectorSSLCA
	})
	commandSet[commands.BOSHCACertCommand] = commands.NewStateQuery(logger, stateValidator, commands.BOSHCACertPropertyName, func(state storage.State) string {
		fmt.Fprintln(os.Stderr, "'bosh-ca-cert' has been deprecated and will be removed in future versions of bbl, please use 'director-ca-cert'")
		return state.BOSH.DirectorSSLCA
	})
	commandSet[commands.SSHKeyCommand] = commands.NewStateQuery(logger, stateValidator, commands.SSHKeyPropertyName, func(state storage.State) string {
		return state.KeyPair.PrivateKey
	})
	commandSet[commands.EnvIDCommand] = commands.NewStateQuery(logger, stateValidator, commands.EnvIDPropertyName, func(state storage.State) string {
		return state.EnvID
	})
	commandSet[commands.PrintEnvCommand] = commands.NewPrintEnv(logger, stateValidator)

	app := application.New(commandSet, configuration, stateStore, usage)

	err = app.Run()
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "\n\n%s\n", err)
	os.Exit(1)
}
