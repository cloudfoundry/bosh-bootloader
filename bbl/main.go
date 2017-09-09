package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/aws/clientmanager"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation"
	"github.com/cloudfoundry/bosh-bootloader/aws/cloudformation/templates"
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/aws/iam"
	"github.com/cloudfoundry/bosh-bootloader/azure"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/certs"
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/config"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/proxy"
	"github.com/cloudfoundry/bosh-bootloader/stack"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"

	awsapplication "github.com/cloudfoundry/bosh-bootloader/application/aws"
	gcpapplication "github.com/cloudfoundry/bosh-bootloader/application/gcp"
	awscloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/aws"
	azurecloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/azure"
	gcpcloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	awsterraform "github.com/cloudfoundry/bosh-bootloader/terraform/aws"
	azureterraform "github.com/cloudfoundry/bosh-bootloader/terraform/azure"
	gcpterraform "github.com/cloudfoundry/bosh-bootloader/terraform/gcp"
)

var (
	Version     string
	gcpBasePath string
)

func main() {
	newConfig := config.NewConfig(storage.GetState)
	parsedFlags, err := newConfig.Bootstrap(os.Args)
	log.SetFlags(0)
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}

	loadedState := parsedFlags.State

	// Utilities
	envIDGenerator := helpers.NewEnvIDGenerator(rand.Reader)
	envGetter := helpers.NewEnvGetter()
	logger := application.NewLogger(os.Stdout)
	stderrLogger := application.NewLogger(os.Stderr)

	// Usage Command
	usage := commands.NewUsage(logger)

	storage.GetStateLogger = stderrLogger

	stateStore := storage.NewStore(parsedFlags.StateDir)
	stateValidator := application.NewStateValidator(parsedFlags.StateDir)

	// Amazon
	awsConfiguration := aws.Config{
		AccessKeyID:     loadedState.AWS.AccessKeyID,
		SecretAccessKey: loadedState.AWS.SecretAccessKey,
		Region:          loadedState.AWS.Region,
	}

	awsClientProvider := &clientmanager.ClientProvider{}
	awsClientProvider.SetConfig(awsConfiguration)

	vpcStatusChecker := ec2.NewVPCStatusChecker(awsClientProvider)
	awsAvailabilityZoneRetriever := ec2.NewAvailabilityZoneRetriever(awsClientProvider)
	templateBuilder := templates.NewTemplateBuilder(logger)
	stackManager := cloudformation.NewStackManager(awsClientProvider, logger)
	infrastructureManager := cloudformation.NewInfrastructureManager(templateBuilder, stackManager)
	certificateDescriber := iam.NewCertificateDescriber(awsClientProvider)
	certificateDeleter := iam.NewCertificateDeleter(awsClientProvider)
	certificateValidator := certs.NewValidator()
	userPolicyDeleter := iam.NewUserPolicyDeleter(awsClientProvider)
	awsKeyPairDeleter := ec2.NewKeyPair(awsClientProvider, logger)

	// GCP
	gcpClientProvider := gcp.NewClientProvider(gcpBasePath)
	if loadedState.IAAS == "gcp" {
		err = gcpClientProvider.SetConfig(loadedState.GCP.ServiceAccountKey, loadedState.GCP.ProjectID, loadedState.GCP.Region, loadedState.GCP.Zone)
		if err != nil {
			log.Fatalf("\n\n%s\n", err)
		}
	}
	gcpNetworkInstancesChecker := gcp.NewNetworkInstancesChecker(gcpClientProvider.Client())
	// EnvID

	envIDManager := helpers.NewEnvIDManager(envIDGenerator, gcpClientProvider.Client(), infrastructureManager)

	// Terraform
	terraformOutputBuffer := bytes.NewBuffer([]byte{})

	terraformCmd := terraform.NewCmd(os.Stderr, terraformOutputBuffer)
	terraformExecutor := terraform.NewExecutor(terraformCmd, parsedFlags.Debug)

	gcpTemplateGenerator := gcpterraform.NewTemplateGenerator()
	gcpInputGenerator := gcpterraform.NewInputGenerator()
	gcpOutputGenerator := gcpterraform.NewOutputGenerator(terraformExecutor)

	awsTemplateGenerator := awsterraform.NewTemplateGenerator()
	awsInputGenerator := awsterraform.NewInputGenerator(awsAvailabilityZoneRetriever)
	awsOutputGenerator := awsterraform.NewOutputGenerator(terraformExecutor)

	azureTemplateGenerator := azureterraform.NewTemplateGenerator()
	azureInputGenerator := azureterraform.NewInputGenerator()
	azureOutputGenerator := azureterraform.NewOutputGenerator(terraformExecutor)

	templateGenerator := terraform.NewTemplateGenerator(gcpTemplateGenerator, awsTemplateGenerator, azureTemplateGenerator)
	inputGenerator := terraform.NewInputGenerator(gcpInputGenerator, awsInputGenerator, azureInputGenerator)
	stackMigrator := stack.NewMigrator(terraformExecutor, infrastructureManager, certificateDescriber, userPolicyDeleter, awsAvailabilityZoneRetriever, awsKeyPairDeleter)

	terraformManager := terraform.NewManager(terraform.NewManagerArgs{
		Executor:              terraformExecutor,
		TemplateGenerator:     templateGenerator,
		InputGenerator:        inputGenerator,
		AWSOutputGenerator:    awsOutputGenerator,
		AzureOutputGenerator:  azureOutputGenerator,
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
	boshClientProvider := bosh.NewClientProvider(socks5Proxy)

	// Environment Validators
	awsBrokenEnvironmentValidator := awsapplication.NewBrokenEnvironmentValidator(infrastructureManager)
	awsEnvironmentValidator := awsapplication.NewEnvironmentValidator(infrastructureManager, boshClientProvider)
	gcpEnvironmentValidator := gcpapplication.NewEnvironmentValidator(boshClientProvider)

	// Cloud Config
	sshKeyGetter := bosh.NewSSHKeyGetter()
	awsCloudFormationOpsGenerator := awscloudconfig.NewCloudFormationOpsGenerator(awsAvailabilityZoneRetriever, infrastructureManager)
	awsTerraformOpsGenerator := awscloudconfig.NewTerraformOpsGenerator(terraformManager)
	gcpOpsGenerator := gcpcloudconfig.NewOpsGenerator(terraformManager)
	azureOpsGenerator := azurecloudconfig.NewOpsGenerator(terraformManager)
	cloudConfigOpsGenerator := cloudconfig.NewOpsGenerator(awsCloudFormationOpsGenerator, awsTerraformOpsGenerator, gcpOpsGenerator, azureOpsGenerator)
	cloudConfigManager := cloudconfig.NewManager(logger, boshCommand, cloudConfigOpsGenerator, boshClientProvider, socks5Proxy, terraformManager, sshKeyGetter)

	// Subcommands
	awsUp := commands.NewAWSUp(boshManager, cloudConfigManager, stateStore, awsClientProvider, envIDManager, terraformManager, awsBrokenEnvironmentValidator)
	awsCreateLBs := commands.NewAWSCreateLBs(cloudConfigManager, stateStore, terraformManager, awsEnvironmentValidator)
	awsLBs := commands.NewAWSLBs(terraformManager, logger)
	awsUpdateLBs := commands.NewAWSUpdateLBs(awsCreateLBs)
	awsDeleteLBs := commands.NewAWSDeleteLBs(logger, cloudConfigManager, stateStore, awsEnvironmentValidator, terraformManager)

	azureClient := azure.NewClient()
	azureUp := commands.NewAzureUp(azureClient, boshManager, cloudConfigManager, envIDManager, logger, stateStore, terraformManager)

	gcpDeleteLBs := commands.NewGCPDeleteLBs(stateStore, gcpEnvironmentValidator, terraformManager, cloudConfigManager)

	gcpUp := commands.NewGCPUp(commands.NewGCPUpArgs{
		StateStore:                   stateStore,
		TerraformManager:             terraformManager,
		BoshManager:                  boshManager,
		Logger:                       logger,
		EnvIDManager:                 envIDManager,
		CloudConfigManager:           cloudConfigManager,
		GCPAvailabilityZoneRetriever: gcpClientProvider.Client(),
	})

	gcpCreateLBs := commands.NewGCPCreateLBs(terraformManager, cloudConfigManager, stateStore, gcpEnvironmentValidator, gcpClientProvider.Client())

	gcpLBs := commands.NewGCPLBs(terraformManager, logger)

	gcpUpdateLBs := commands.NewGCPUpdateLBs(gcpCreateLBs)

	up := commands.NewUp(awsUp, gcpUp, azureUp, envGetter, boshManager)

	// Commands
	commandSet := application.CommandSet{}
	commandSet["help"] = usage
	commandSet["version"] = commands.NewVersion(Version, logger)
	commandSet["up"] = up
	sshKeyDeleter := bosh.NewSSHKeyDeleter()
	commandSet["rotate"] = commands.NewRotate(stateValidator, sshKeyDeleter, up)
	commandSet["destroy"] = commands.NewDestroy(logger, os.Stdin, boshManager, vpcStatusChecker, stackManager, infrastructureManager, certificateDeleter, stateStore, stateValidator, terraformManager, gcpNetworkInstancesChecker)
	commandSet["down"] = commandSet["destroy"]
	commandSet["create-lbs"] = commands.NewCreateLBs(awsCreateLBs, gcpCreateLBs, logger, stateValidator, certificateValidator, boshManager)
	commandSet["update-lbs"] = commands.NewUpdateLBs(awsUpdateLBs, gcpUpdateLBs, certificateValidator, stateValidator, logger, boshManager)
	commandSet["delete-lbs"] = commands.NewDeleteLBs(gcpDeleteLBs, awsDeleteLBs, logger, stateValidator, boshManager)
	commandSet["lbs"] = commands.NewLBs(gcpLBs, awsLBs, stateValidator, logger)
	commandSet["jumpbox-address"] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.JumpboxAddressPropertyName)
	commandSet["director-address"] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorAddressPropertyName)
	commandSet["director-username"] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorUsernamePropertyName)
	commandSet["director-password"] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorPasswordPropertyName)
	commandSet["director-ca-cert"] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.DirectorCACertPropertyName)
	commandSet["ssh-key"] = commands.NewSSHKey(logger, stateValidator, sshKeyGetter)
	commandSet["env-id"] = commands.NewStateQuery(logger, stateValidator, terraformManager, infrastructureManager, commands.EnvIDPropertyName)
	commandSet["latest-error"] = commands.NewLatestError(logger, stateValidator)
	commandSet["print-env"] = commands.NewPrintEnv(logger, stateValidator, terraformManager)
	commandSet["cloud-config"] = commands.NewCloudConfig(logger, stateValidator, cloudConfigManager)
	commandSet["bosh-deployment-vars"] = commands.NewBOSHDeploymentVars(logger, boshManager, stateValidator, terraformManager)

	commandConfiguration := &application.Configuration{
		Global: application.GlobalConfiguration{
			StateDir: parsedFlags.StateDir,
			Debug:    parsedFlags.Debug,
		},
		State:           loadedState,
		ShowCommandHelp: parsedFlags.Help,
	}

	if len(parsedFlags.RemainingArgs) > 0 {
		commandConfiguration.Command = parsedFlags.RemainingArgs[0]
		commandConfiguration.SubcommandFlags = parsedFlags.RemainingArgs[1:]
	} else {
		commandConfiguration.ShowCommandHelp = false
		if parsedFlags.Help {
			commandConfiguration.Command = "help"
		}
		if parsedFlags.Version {
			commandConfiguration.Command = "version"
		}
	}

	if len(os.Args) == 1 {
		commandConfiguration.Command = "help"
	}

	app := application.New(commandSet, *commandConfiguration, usage)

	err = app.Run()
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}
}
