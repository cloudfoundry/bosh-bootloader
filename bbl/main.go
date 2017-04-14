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
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	awscloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/aws"
	gcpcloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	gcpterraform "github.com/cloudfoundry/bosh-bootloader/terraform/gcp"
)

var (
	Version     string
	gcpBasePath string
)

func main() {
	// Command Set
	commandSet := application.CommandSet{
		commands.HelpCommand:               nil,
		commands.VersionCommand:            nil,
		commands.UpCommand:                 nil,
		commands.DestroyCommand:            nil,
		commands.DirectorAddressCommand:    nil,
		commands.DirectorUsernameCommand:   nil,
		commands.DirectorPasswordCommand:   nil,
		commands.DirectorCACertCommand:     nil,
		commands.SSHKeyCommand:             nil,
		commands.CreateLBsCommand:          nil,
		commands.UpdateLBsCommand:          nil,
		commands.DeleteLBsCommand:          nil,
		commands.LBsCommand:                nil,
		commands.EnvIDCommand:              nil,
		commands.PrintEnvCommand:           nil,
		commands.CloudConfigCommand:        nil,
		commands.BOSHDeploymentVarsCommand: nil,
	}

	// Utilities
	uuidGenerator := helpers.NewUUIDGenerator(rand.Reader)
	stringGenerator := helpers.NewStringGenerator(rand.Reader)
	envIDGenerator := helpers.NewEnvIDGenerator(rand.Reader)
	logger := application.NewLogger(os.Stdout)
	stderrLogger := application.NewLogger(os.Stderr)

	// Usage Command
	usage := commands.NewUsage(os.Stdout)

	configuration := getConfiguration(usage.Print, commandSet)

	storage.GetStateLogger = stderrLogger

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
	gcpKeyPairDeleter := gcp.NewKeyPairDeleter(gcpClientProvider, logger)
	gcpNetworkInstancesChecker := gcp.NewNetworkInstancesChecker(gcpClientProvider)
	zones := gcp.NewZones()

	// EnvID
	envIDManager := helpers.NewEnvIDManager(envIDGenerator, gcpClientProvider, infrastructureManager)

	// Terraform
	terraformCmd := terraform.NewCmd(os.Stderr)
	terraformExecutor := terraform.NewExecutor(terraformCmd, configuration.Global.Debug)
	gcpTemplateGenerator := gcpterraform.NewTemplateGenerator(zones)
	gcpInputGenerator := gcpterraform.NewInputGenerator()
	gcpOutputGenerator := gcpterraform.NewOutputGenerator(terraformExecutor)
	templateGenerator := terraform.NewTemplateGenerator(gcpTemplateGenerator)
	terraformManager := terraform.NewManager(terraformExecutor, templateGenerator, gcpInputGenerator, gcpOutputGenerator, logger)

	// BOSH
	boshCommand := bosh.NewCmd(os.Stderr)
	boshExecutor := bosh.NewExecutor(boshCommand, ioutil.TempDir, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal,
		json.Marshal, ioutil.WriteFile)
	boshManager := bosh.NewManager(boshExecutor, terraformManager, stackManager, logger)
	boshClientProvider := bosh.NewClientProvider()

	// Cloud Config
	awsOpsGenerator := awscloudconfig.NewOpsGenerator(availabilityZoneRetriever, infrastructureManager)
	gcpOpsGenerator := gcpcloudconfig.NewOpsGenerator(terraformManager, zones)
	cloudConfigOpsGenerator := cloudconfig.NewOpsGenerator(awsOpsGenerator, gcpOpsGenerator)
	cloudConfigManager := cloudconfig.NewManager(logger, boshCommand, cloudConfigOpsGenerator, boshClientProvider)

	// Subcommands
	awsUp := commands.NewAWSUp(
		credentialValidator, infrastructureManager, keyPairSynchronizer, boshManager,
		availabilityZoneRetriever, certificateDescriber,
		cloudConfigManager, stateStore, clientProvider, envIDManager)

	awsCreateLBs := commands.NewAWSCreateLBs(
		logger, credentialValidator, certificateManager, infrastructureManager,
		availabilityZoneRetriever, boshClientProvider, cloudConfigManager, certificateValidator,
		uuidGenerator, stateStore,
	)

	awsUpdateLBs := commands.NewAWSUpdateLBs(credentialValidator, certificateManager, availabilityZoneRetriever, infrastructureManager,
		boshClientProvider, logger, uuidGenerator, stateStore)

	awsDeleteLBs := commands.NewAWSDeleteLBs(
		credentialValidator, availabilityZoneRetriever, certificateManager,
		infrastructureManager, logger, cloudConfigManager, boshClientProvider, stateStore,
	)
	gcpDeleteLBs := commands.NewGCPDeleteLBs(stateStore, terraformManager, cloudConfigManager)

	gcpUp := commands.NewGCPUp(commands.NewGCPUpArgs{
		StateStore:         stateStore,
		KeyPairUpdater:     gcpKeyPairUpdater,
		GCPProvider:        gcpClientProvider,
		TerraformManager:   terraformManager,
		BoshManager:        boshManager,
		Logger:             logger,
		EnvIDManager:       envIDManager,
		CloudConfigManager: cloudConfigManager,
	})

	gcpCreateLBs := commands.NewGCPCreateLBs(terraformManager, boshClientProvider, cloudConfigManager, stateStore, logger)

	gcpUpdateLBs := commands.NewGCPUpdateLBs(gcpCreateLBs)

	envGetter := commands.NewEnvGetter()

	// Commands
	commandSet[commands.HelpCommand] = commands.NewUsage(os.Stdout)
	commandSet[commands.VersionCommand] = commands.NewVersion(Version, os.Stdout)
	commandSet[commands.UpCommand] = commands.NewUp(awsUp, gcpUp, envGetter, boshManager)
	commandSet[commands.DestroyCommand] = commands.NewDestroy(
		credentialValidator, logger, os.Stdin, boshManager, vpcStatusChecker, stackManager,
		stringGenerator, infrastructureManager, awsKeyPairDeleter, gcpKeyPairDeleter, certificateDeleter,
		stateStore, stateValidator, terraformManager, gcpNetworkInstancesChecker,
	)
	commandSet[commands.CreateLBsCommand] = commands.NewCreateLBs(awsCreateLBs, gcpCreateLBs, stateValidator, boshManager)
	commandSet[commands.UpdateLBsCommand] = commands.NewUpdateLBs(awsUpdateLBs, gcpUpdateLBs, certificateValidator, stateValidator, logger, boshManager)
	commandSet[commands.DeleteLBsCommand] = commands.NewDeleteLBs(gcpDeleteLBs, awsDeleteLBs, logger, stateValidator, boshManager)
	commandSet[commands.LBsCommand] = commands.NewLBs(credentialValidator, stateValidator, infrastructureManager, terraformManager, os.Stdout)
	commandSet[commands.DirectorAddressCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorAddressPropertyName)
	commandSet[commands.DirectorUsernameCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorUsernamePropertyName)
	commandSet[commands.DirectorPasswordCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorPasswordPropertyName)
	commandSet[commands.DirectorCACertCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorCACertPropertyName)
	commandSet[commands.SSHKeyCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.SSHKeyPropertyName)
	commandSet[commands.EnvIDCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.EnvIDPropertyName)
	commandSet[commands.PrintEnvCommand] = commands.NewPrintEnv(logger, stateValidator, terraformManager, infrastructureManager)
	commandSet[commands.CloudConfigCommand] = commands.NewCloudConfig(logger, stateValidator, cloudConfigManager)
	commandSet[commands.BOSHDeploymentVarsCommand] = commands.NewBOSHDeploymentVars(logger, boshManager)

	app := application.New(commandSet, configuration, stateStore, usage)

	err := app.Run()
	if err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "\n\n%s\n", err)
	os.Exit(1)
}

func getConfiguration(printUsage func(), commandSet application.CommandSet) application.Configuration {
	commandLineParser := application.NewCommandLineParser(printUsage, commandSet)
	configurationParser := application.NewConfigurationParser(commandLineParser)
	configuration, err := configurationParser.Parse(os.Args[1:])
	if err != nil {
		fail(err)
	}

	return configuration
}
