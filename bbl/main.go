package main

import (
	"bytes"
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
	"github.com/cloudfoundry/bosh-bootloader/keypair"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	awsapplication "github.com/cloudfoundry/bosh-bootloader/application/aws"
	gcpapplication "github.com/cloudfoundry/bosh-bootloader/application/gcp"
	awscloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/aws"
	gcpcloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	awskeypair "github.com/cloudfoundry/bosh-bootloader/keypair/aws"
	gcpkeypair "github.com/cloudfoundry/bosh-bootloader/keypair/gcp"
	awsterraform "github.com/cloudfoundry/bosh-bootloader/terraform/aws"
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
		commands.DownCommand:               nil,
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
		commands.LatestErrorCommand:        nil,
		commands.PrintEnvCommand:           nil,
		commands.CloudConfigCommand:        nil,
		commands.BOSHDeploymentVarsCommand: nil,
	}

	// Utilities
	uuidGenerator := helpers.NewUUIDGenerator(rand.Reader)
	stringGenerator := helpers.NewStringGenerator(rand.Reader)
	envIDGenerator := helpers.NewEnvIDGenerator(rand.Reader)
	envGetter := helpers.NewEnvGetter()
	logger := application.NewLogger(os.Stdout)
	stderrLogger := application.NewLogger(os.Stderr)

	// Usage Command
	usage := commands.NewUsage(logger)

	configuration := getConfiguration(usage.Print, commandSet, envGetter)

	storage.GetStateLogger = stderrLogger

	stateStore := storage.NewStore(configuration.Global.StateDir)
	stateValidator := application.NewStateValidator(configuration.Global.StateDir)

	awsCredentialValidator := awsapplication.NewCredentialValidator(configuration)
	gcpCredentialValidator := gcpapplication.NewCredentialValidator(configuration)
	credentialValidator := application.NewCredentialValidator(configuration, gcpCredentialValidator, awsCredentialValidator)

	// Amazon
	awsConfiguration := aws.Config{
		AccessKeyID:      configuration.State.AWS.AccessKeyID,
		SecretAccessKey:  configuration.State.AWS.SecretAccessKey,
		Region:           configuration.State.AWS.Region,
		EndpointOverride: configuration.Global.EndpointOverride,
	}

	clientProvider := &clientmanager.ClientProvider{EndpointOverride: configuration.Global.EndpointOverride}
	clientProvider.SetConfig(awsConfiguration)

	vpcStatusChecker := ec2.NewVPCStatusChecker(clientProvider)
	awsKeyPairCreator := ec2.NewKeyPairCreator(clientProvider)
	awsKeyPairDeleter := ec2.NewKeyPairDeleter(clientProvider, logger)
	keyPairChecker := ec2.NewKeyPairChecker(clientProvider)
	keyPairSynchronizer := ec2.NewKeyPairSynchronizer(awsKeyPairCreator, keyPairChecker, logger)
	awsKeyPairManager := awskeypair.NewManager(keyPairSynchronizer)
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
	gcpKeyPairManager := gcpkeypair.NewManager(gcpKeyPairUpdater)
	zones := gcp.NewZones()

	// EnvID
	envIDManager := helpers.NewEnvIDManager(envIDGenerator, gcpClientProvider, infrastructureManager)

	// Keypair Manager
	keyPairManager := keypair.NewManager(awsKeyPairManager, gcpKeyPairManager)

	// Terraform
	terraformOutputBuffer := bytes.NewBuffer([]byte{})

	terraformCmd := terraform.NewCmd(os.Stderr, terraformOutputBuffer)
	terraformExecutor := terraform.NewExecutor(terraformCmd, configuration.Global.Debug)
	gcpTemplateGenerator := gcpterraform.NewTemplateGenerator(zones)
	gcpInputGenerator := gcpterraform.NewInputGenerator()
	gcpOutputGenerator := gcpterraform.NewOutputGenerator(terraformExecutor)
	awsTemplateGenerator := awsterraform.NewTemplateGenerator()
	awsInputGenerator := awsterraform.NewInputGenerator(availabilityZoneRetriever)
	awsOutputGenerator := awsterraform.NewOutputGenerator(terraformExecutor)
	templateGenerator := terraform.NewTemplateGenerator(gcpTemplateGenerator, awsTemplateGenerator)
	inputGenerator := terraform.NewInputGenerator(gcpInputGenerator, awsInputGenerator)
	outputGenerator := terraform.NewOutputGenerator(gcpOutputGenerator, awsOutputGenerator)
	terraformManager := terraform.NewManager(terraform.NewManagerArgs{
		Executor:              terraformExecutor,
		TemplateGenerator:     templateGenerator,
		InputGenerator:        inputGenerator,
		OutputGenerator:       outputGenerator,
		TerraformOutputBuffer: terraformOutputBuffer,
		Logger:                logger,
	})

	// BOSH
	boshCommand := bosh.NewCmd(os.Stderr)
	boshExecutor := bosh.NewExecutor(boshCommand, ioutil.TempDir, ioutil.ReadFile, yaml.Unmarshal, json.Unmarshal,
		json.Marshal, ioutil.WriteFile)
	boshManager := bosh.NewManager(boshExecutor, terraformManager, stackManager, logger)
	boshClientProvider := bosh.NewClientProvider()

	// Environment Validators
	awsBrokenEnvironmentValidator := awsapplication.NewBrokenEnvironmentValidator(infrastructureManager)
	awsEnvironmentValidator := awsapplication.NewEnvironmentValidator(infrastructureManager, boshClientProvider)
	gcpEnvironmentValidator := gcpapplication.NewEnvironmentValidator(boshClientProvider)

	// Cloud Config
	awsCloudFormationOpsGenerator := awscloudconfig.NewCloudFormationOpsGenerator(availabilityZoneRetriever, infrastructureManager)
	awsTerraformOpsGenerator := awscloudconfig.NewTerraformOpsGenerator(availabilityZoneRetriever, terraformManager)
	gcpOpsGenerator := gcpcloudconfig.NewOpsGenerator(terraformManager, zones)
	cloudConfigOpsGenerator := cloudconfig.NewOpsGenerator(awsCloudFormationOpsGenerator, awsTerraformOpsGenerator, gcpOpsGenerator)
	cloudConfigManager := cloudconfig.NewManager(logger, boshCommand, cloudConfigOpsGenerator, boshClientProvider)

	// Subcommands
	awsUp := commands.NewAWSUp(
		awsCredentialValidator, infrastructureManager, keyPairManager, boshManager,
		availabilityZoneRetriever, certificateDescriber,
		cloudConfigManager, stateStore, clientProvider, envIDManager, terraformManager, awsBrokenEnvironmentValidator)

	awsCreateLBs := commands.NewAWSCreateLBs(
		logger, awsCredentialValidator, certificateManager, infrastructureManager,
		availabilityZoneRetriever, cloudConfigManager, certificateValidator,
		uuidGenerator, stateStore, terraformManager, awsEnvironmentValidator,
	)

	awsUpdateLBs := commands.NewAWSUpdateLBs(awsCredentialValidator, certificateManager, availabilityZoneRetriever, infrastructureManager,
		logger, uuidGenerator, stateStore, awsEnvironmentValidator)

	awsDeleteLBs := commands.NewAWSDeleteLBs(
		awsCredentialValidator, availabilityZoneRetriever, certificateManager,
		infrastructureManager, logger, cloudConfigManager, stateStore, awsEnvironmentValidator,
		terraformManager,
	)
	gcpDeleteLBs := commands.NewGCPDeleteLBs(stateStore, terraformManager, cloudConfigManager)

	gcpUp := commands.NewGCPUp(commands.NewGCPUpArgs{
		StateStore:         stateStore,
		KeyPairManager:     keyPairManager,
		GCPProvider:        gcpClientProvider,
		TerraformManager:   terraformManager,
		BoshManager:        boshManager,
		Logger:             logger,
		EnvIDManager:       envIDManager,
		CloudConfigManager: cloudConfigManager,
	})

	gcpCreateLBs := commands.NewGCPCreateLBs(terraformManager, cloudConfigManager, stateStore, logger, gcpEnvironmentValidator)

	gcpUpdateLBs := commands.NewGCPUpdateLBs(gcpCreateLBs)

	// Commands
	commandSet[commands.HelpCommand] = usage
	commandSet[commands.VersionCommand] = commands.NewVersion(Version, logger)
	commandSet[commands.UpCommand] = commands.NewUp(awsUp, gcpUp, envGetter, boshManager)
	commandSet[commands.DestroyCommand] = commands.NewDestroy(
		credentialValidator, logger, os.Stdin, boshManager, vpcStatusChecker, stackManager,
		stringGenerator, infrastructureManager, awsKeyPairDeleter, gcpKeyPairDeleter, certificateDeleter,
		stateStore, stateValidator, terraformManager, gcpNetworkInstancesChecker,
	)
	commandSet[commands.DownCommand] = commandSet[commands.DestroyCommand]
	commandSet[commands.CreateLBsCommand] = commands.NewCreateLBs(awsCreateLBs, gcpCreateLBs, stateValidator, boshManager)
	commandSet[commands.UpdateLBsCommand] = commands.NewUpdateLBs(awsUpdateLBs, gcpUpdateLBs, certificateValidator, stateValidator, logger, boshManager)
	commandSet[commands.DeleteLBsCommand] = commands.NewDeleteLBs(gcpDeleteLBs, awsDeleteLBs, logger, stateValidator, boshManager)
	commandSet[commands.LBsCommand] = commands.NewLBs(awsCredentialValidator, stateValidator, infrastructureManager, terraformManager, logger)
	commandSet[commands.DirectorAddressCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorAddressPropertyName)
	commandSet[commands.DirectorUsernameCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorUsernamePropertyName)
	commandSet[commands.DirectorPasswordCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorPasswordPropertyName)
	commandSet[commands.DirectorCACertCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorCACertPropertyName)
	commandSet[commands.SSHKeyCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.SSHKeyPropertyName)
	commandSet[commands.EnvIDCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.EnvIDPropertyName)
	commandSet[commands.LatestErrorCommand] = commands.NewLatestError(logger)
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

func getConfiguration(printUsage func(), commandSet application.CommandSet, envGetter helpers.EnvGetter) application.Configuration {
	commandLineParser := application.NewCommandLineParser(printUsage, commandSet, envGetter)
	configurationParser := application.NewConfigurationParser(commandLineParser)
	configuration, err := configurationParser.Parse(os.Args[1:])
	if err != nil {
		fail(err)
	}

	return configuration
}
