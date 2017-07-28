package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"golang.org/x/crypto/ssh"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/clientmanager"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/certs"
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/keypair"
	"github.com/cloudfoundry/bosh-bootloader/proxy"
	"github.com/cloudfoundry/bosh-bootloader/stack"
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
		commands.JumpboxAddressCommand:     nil,
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
		commands.RotateCommand:             nil,
	}

	// Utilities
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
	awsKeyPairManager := awskeypair.NewManager(keyPairSynchronizer, awsKeyPairDeleter, clientProvider)
	awsAvailabilityZoneRetriever := ec2.NewAvailabilityZoneRetriever(clientProvider)
	templateBuilder := templates.NewTemplateBuilder(logger)
	stackManager := cloudformation.NewStackManager(clientProvider, logger)
	infrastructureManager := cloudformation.NewInfrastructureManager(templateBuilder, stackManager)
	certificateDescriber := iam.NewCertificateDescriber(clientProvider)
	certificateDeleter := iam.NewCertificateDeleter(clientProvider)
	certificateValidator := certs.NewValidator()
	userPolicyDeleter := iam.NewUserPolicyDeleter(clientProvider)

	// GCP
	gcpClientProvider := gcp.NewClientProvider(gcpBasePath)
	gcpClientProvider.SetConfig(configuration.State.GCP.ServiceAccountKey, configuration.State.GCP.ProjectID, configuration.State.GCP.Region, configuration.State.GCP.Zone)
	gcpKeyPairUpdater := gcp.NewKeyPairUpdater(rand.Reader, rsa.GenerateKey, ssh.NewPublicKey, gcpClientProvider, logger)
	gcpKeyPairDeleter := gcp.NewKeyPairDeleter(gcpClientProvider, logger)
	gcpNetworkInstancesChecker := gcp.NewNetworkInstancesChecker(gcpClientProvider)
	gcpKeyPairManager := gcpkeypair.NewManager(gcpKeyPairUpdater, gcpKeyPairDeleter, gcpClientProvider)
	gcpAvailabilityZoneRetriever := gcp.NewZones(gcpClientProvider)

	// EnvID
	envIDManager := helpers.NewEnvIDManager(envIDGenerator, gcpClientProvider, infrastructureManager)

	// Keypair Manager
	keyPairManager := keypair.NewManager(awsKeyPairManager, gcpKeyPairManager)

	// Terraform
	terraformOutputBuffer := bytes.NewBuffer([]byte{})

	terraformCmd := terraform.NewCmd(os.Stderr, terraformOutputBuffer)
	terraformExecutor := terraform.NewExecutor(terraformCmd, configuration.Global.Debug)
	gcpTemplateGenerator := gcpterraform.NewTemplateGenerator()
	gcpInputGenerator := gcpterraform.NewInputGenerator()
	gcpOutputGenerator := gcpterraform.NewOutputGenerator(terraformExecutor)
	awsTemplateGenerator := awsterraform.NewTemplateGenerator()
	awsInputGenerator := awsterraform.NewInputGenerator(awsAvailabilityZoneRetriever)
	awsOutputGenerator := awsterraform.NewOutputGenerator(terraformExecutor)
	templateGenerator := terraform.NewTemplateGenerator(gcpTemplateGenerator, awsTemplateGenerator)
	inputGenerator := terraform.NewInputGenerator(gcpInputGenerator, awsInputGenerator)
	stackMigrator := stack.NewMigrator(terraformExecutor, infrastructureManager, certificateDescriber, userPolicyDeleter, awsAvailabilityZoneRetriever)
	terraformManager := terraform.NewManager(terraform.NewManagerArgs{
		Executor:              terraformExecutor,
		TemplateGenerator:     templateGenerator,
		InputGenerator:        inputGenerator,
		AWSOutputGenerator:    awsOutputGenerator,
		GCPOutputGenerator:    gcpOutputGenerator,
		TerraformOutputBuffer: terraformOutputBuffer,
		Logger:                logger,
		StackMigrator:         stackMigrator,
	})

	// BOSH
	hostKeyGetter := proxy.NewHostKeyGetter()
	socks5Proxy := proxy.NewSocks5Proxy(logger, hostKeyGetter, 0)
	boshCommand := bosh.NewCmd(os.Stderr)
	boshExecutor := bosh.NewExecutor(boshCommand, ioutil.TempDir, ioutil.ReadFile, json.Unmarshal,
		json.Marshal, ioutil.WriteFile)
	boshManager := bosh.NewManager(boshExecutor, logger, socks5Proxy)
	boshClientProvider := bosh.NewClientProvider()

	// Environment Validators
	awsBrokenEnvironmentValidator := awsapplication.NewBrokenEnvironmentValidator(infrastructureManager)
	awsEnvironmentValidator := awsapplication.NewEnvironmentValidator(infrastructureManager, boshClientProvider)
	gcpEnvironmentValidator := gcpapplication.NewEnvironmentValidator(boshClientProvider)

	// Cloud Config
	sshKeyGetter := bosh.NewSSHKeyGetter()
	awsCloudFormationOpsGenerator := awscloudconfig.NewCloudFormationOpsGenerator(awsAvailabilityZoneRetriever, infrastructureManager)
	awsTerraformOpsGenerator := awscloudconfig.NewTerraformOpsGenerator(terraformManager)
	gcpOpsGenerator := gcpcloudconfig.NewOpsGenerator(terraformManager)
	cloudConfigOpsGenerator := cloudconfig.NewOpsGenerator(awsCloudFormationOpsGenerator, awsTerraformOpsGenerator, gcpOpsGenerator)
	cloudConfigManager := cloudconfig.NewManager(logger, boshCommand, cloudConfigOpsGenerator, boshClientProvider, socks5Proxy, terraformManager, sshKeyGetter)

	// Subcommands
	awsUp := commands.NewAWSUp(
		awsCredentialValidator, keyPairManager, boshManager,
		cloudConfigManager, stateStore, clientProvider, envIDManager, terraformManager, awsBrokenEnvironmentValidator)

	awsCreateLBs := commands.NewAWSCreateLBs(
		logger, awsCredentialValidator, cloudConfigManager,
		stateStore, terraformManager, awsEnvironmentValidator,
	)

	awsLBs := commands.NewAWSLBs(terraformManager, logger)

	awsUpdateLBs := commands.NewAWSUpdateLBs(awsCreateLBs, awsCredentialValidator, awsEnvironmentValidator)

	awsDeleteLBs := commands.NewAWSDeleteLBs(
		awsCredentialValidator, logger, cloudConfigManager, stateStore, awsEnvironmentValidator,
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

	gcpCreateLBs := commands.NewGCPCreateLBs(terraformManager, cloudConfigManager, stateStore, logger, gcpEnvironmentValidator, gcpAvailabilityZoneRetriever)

	gcpLBs := commands.NewGCPLBs(terraformManager, logger)

	gcpUpdateLBs := commands.NewGCPUpdateLBs(gcpCreateLBs)

	// Commands
	commandSet[commands.HelpCommand] = usage
	commandSet[commands.VersionCommand] = commands.NewVersion(Version, logger)
	commandSet[commands.UpCommand] = commands.NewUp(awsUp, gcpUp, envGetter, boshManager)
	commandSet[commands.DestroyCommand] = commands.NewDestroy(
		credentialValidator, logger, os.Stdin, boshManager, vpcStatusChecker, stackManager,
		infrastructureManager, awsKeyPairDeleter, gcpKeyPairDeleter, certificateDeleter,
		stateStore, stateValidator, terraformManager, gcpNetworkInstancesChecker,
	)
	commandSet[commands.DownCommand] = commandSet[commands.DestroyCommand]
	commandSet[commands.CreateLBsCommand] = commands.NewCreateLBs(awsCreateLBs, gcpCreateLBs, stateValidator, certificateValidator, boshManager)
	commandSet[commands.UpdateLBsCommand] = commands.NewUpdateLBs(awsUpdateLBs, gcpUpdateLBs, certificateValidator, stateValidator, logger, boshManager)
	commandSet[commands.DeleteLBsCommand] = commands.NewDeleteLBs(gcpDeleteLBs, awsDeleteLBs, logger, stateValidator, boshManager)
	commandSet[commands.LBsCommand] = commands.NewLBs(gcpLBs, awsLBs, stateValidator, logger)
	commandSet[commands.JumpboxAddressCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.JumpboxAddressPropertyName)
	commandSet[commands.DirectorAddressCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorAddressPropertyName)
	commandSet[commands.DirectorUsernameCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorUsernamePropertyName)
	commandSet[commands.DirectorPasswordCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorPasswordPropertyName)
	commandSet[commands.DirectorCACertCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorCACertPropertyName)
	commandSet[commands.SSHKeyCommand] = commands.NewSSHKey(logger, stateValidator, sshKeyGetter)
	commandSet[commands.EnvIDCommand] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.EnvIDPropertyName)
	commandSet[commands.LatestErrorCommand] = commands.NewLatestError(logger, stateValidator)
	commandSet[commands.PrintEnvCommand] = commands.NewPrintEnv(logger, stateValidator, terraformManager)
	commandSet[commands.CloudConfigCommand] = commands.NewCloudConfig(logger, stateValidator, cloudConfigManager)
	commandSet[commands.BOSHDeploymentVarsCommand] = commands.NewBOSHDeploymentVars(logger, boshManager, stateValidator, terraformManager)
	commandSet[commands.RotateCommand] = commands.NewRotate(stateStore, keyPairManager, terraformManager, boshManager, stateValidator)

	app := application.New(commandSet, configuration, stateStore, usage)

	err := app.Run()
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}
}

func getConfiguration(printUsage func(), commandSet application.CommandSet, envGetter helpers.EnvGetter) application.Configuration {
	commandLineParser := application.NewCommandLineParser(printUsage, commandSet, envGetter)
	configurationParser := application.NewConfigurationParser(commandLineParser)
	configuration, err := configurationParser.Parse(os.Args[1:])
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}

	return configuration
}
