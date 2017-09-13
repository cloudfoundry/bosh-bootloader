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

	appConfig := &application.Configuration{
		Global: application.GlobalConfiguration{
			StateDir: parsedFlags.StateDir,
			Debug:    parsedFlags.Debug,
		},
		State:           parsedFlags.State,
		ShowCommandHelp: parsedFlags.Help,
	}
	if len(parsedFlags.RemainingArgs) > 0 {
		appConfig.Command = parsedFlags.RemainingArgs[0]
		appConfig.SubcommandFlags = parsedFlags.RemainingArgs[1:]
	} else {
		appConfig.ShowCommandHelp = false
		if parsedFlags.Help {
			appConfig.Command = "help"
		}
		if parsedFlags.Version {
			appConfig.Command = "version"
		}
	}
	if len(os.Args) == 1 {
		appConfig.Command = "help"
	}

	needsIAASConfig := config.NeedsIAASConfig(appConfig.Command) && !appConfig.ShowCommandHelp
	if needsIAASConfig {
		err = config.ValidateIAAS(appConfig.State, appConfig.Command)
		if err != nil {
			log.Fatalf("\n\n%s\n", err)
		}
	}

	// Utilities
	envIDGenerator := helpers.NewEnvIDGenerator(rand.Reader)
	logger := application.NewLogger(os.Stdout)
	stderrLogger := application.NewLogger(os.Stderr)

	// Usage Command
	usage := commands.NewUsage(logger)

	storage.GetStateLogger = stderrLogger

	stateStore := storage.NewStore(parsedFlags.StateDir)
	stateValidator := application.NewStateValidator(parsedFlags.StateDir)

	// Terraform
	terraformOutputBuffer := bytes.NewBuffer([]byte{})
	terraformCmd := terraform.NewCmd(os.Stderr, terraformOutputBuffer)
	terraformExecutor := terraform.NewExecutor(terraformCmd, parsedFlags.Debug)

	var (
		stackMigrator                stack.Migrator
		awsAvailabilityZoneRetriever ec2.AvailabilityZoneRetriever
		certificateDeleter           iam.CertificateDeleter
		certificateValidator         certs.Validator
		vpcStatusChecker             ec2.VPCStatusChecker
		infrastructureManager        cloudformation.InfrastructureManager
		stackManager                 cloudformation.StackManager
	)
	awsClientProvider := &clientmanager.ClientProvider{}
	if appConfig.State.IAAS == "aws" && needsIAASConfig {
		awsConfiguration := aws.Config{
			AccessKeyID:     appConfig.State.AWS.AccessKeyID,
			SecretAccessKey: appConfig.State.AWS.SecretAccessKey,
			Region:          appConfig.State.AWS.Region,
		}
		awsClientProvider.SetConfig(awsConfiguration)

		templateBuilder := templates.NewTemplateBuilder(logger)
		certificateDescriber := iam.NewCertificateDescriber(awsClientProvider)
		userPolicyDeleter := iam.NewUserPolicyDeleter(awsClientProvider)
		awsKeyPairDeleter := ec2.NewKeyPair(awsClientProvider, logger)

		awsAvailabilityZoneRetriever = ec2.NewAvailabilityZoneRetriever(awsClientProvider)
		certificateDeleter = iam.NewCertificateDeleter(awsClientProvider)
		certificateValidator = certs.NewValidator()
		vpcStatusChecker = ec2.NewVPCStatusChecker(awsClientProvider)
		infrastructureManager = cloudformation.NewInfrastructureManager(templateBuilder, stackManager)
		stackManager = cloudformation.NewStackManager(awsClientProvider, logger)

		stackMigrator = stack.NewMigrator(terraformExecutor, infrastructureManager, certificateDescriber, userPolicyDeleter, awsAvailabilityZoneRetriever, awsKeyPairDeleter)
	}

	gcpClientProvider := gcp.NewClientProvider(gcpBasePath)
	if appConfig.State.IAAS == "gcp" && needsIAASConfig {
		err = gcpClientProvider.SetConfig(appConfig.State.GCP.ServiceAccountKey, appConfig.State.GCP.ProjectID, appConfig.State.GCP.Region, appConfig.State.GCP.Zone)
		if err != nil {
			log.Fatalf("\n\n%s\n", err)
		}
	}
	gcpNetworkInstancesChecker := gcp.NewNetworkInstancesChecker(gcpClientProvider.Client())

	var envIDManager helpers.EnvIDManager
	if appConfig.State.IAAS != "" {
		envIDManager = helpers.NewEnvIDManager(envIDGenerator, gcpClientProvider.Client(), infrastructureManager, awsClientProvider.GetEC2Client())
	}

	var (
		inputGenerator    terraform.InputGenerator
		outputGenerator   terraform.OutputGenerator
		templateGenerator terraform.TemplateGenerator
	)

	if appConfig.State.IAAS == "aws" {
		templateGenerator = awsterraform.NewTemplateGenerator()
		inputGenerator = awsterraform.NewInputGenerator(awsAvailabilityZoneRetriever)
		outputGenerator = awsterraform.NewOutputGenerator(terraformExecutor)
	} else if appConfig.State.IAAS == "azure" {
		templateGenerator = azureterraform.NewTemplateGenerator()
		inputGenerator = azureterraform.NewInputGenerator()
		outputGenerator = azureterraform.NewOutputGenerator(terraformExecutor)
	} else if appConfig.State.IAAS == "gcp" {
		outputGenerator = gcpterraform.NewOutputGenerator(terraformExecutor)
		templateGenerator = gcpterraform.NewTemplateGenerator()
		inputGenerator = gcpterraform.NewInputGenerator()
	}

	terraformManager := terraform.NewManager(terraform.NewManagerArgs{
		Executor:              terraformExecutor,
		TemplateGenerator:     templateGenerator,
		InputGenerator:        inputGenerator,
		OutputGenerator:       outputGenerator,
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
	var (
		upCmd        commands.UpCmd
		lbsCmd       commands.LBsCmd
		deleteLBsCmd commands.DeleteLBsCmd
	)
	if appConfig.State.IAAS == "aws" {
		upCmd = commands.NewAWSUp(boshManager, cloudConfigManager, stateStore, envIDManager, terraformManager)
		lbsCmd = commands.NewAWSLBs(terraformManager, logger)
		deleteLBsCmd = commands.NewAWSDeleteLBs(cloudConfigManager, stateStore, awsEnvironmentValidator, terraformManager)
	} else if appConfig.State.IAAS == "gcp" {
		upCmd = commands.NewGCPUp(stateStore, terraformManager, boshManager, cloudConfigManager, envIDManager, gcpClientProvider.Client())
		lbsCmd = commands.NewGCPLBs(terraformManager, logger)
		deleteLBsCmd = commands.NewGCPDeleteLBs(stateStore, gcpEnvironmentValidator, terraformManager, cloudConfigManager)
	} else if appConfig.State.IAAS == "azure" {
		azureClient := azure.NewClient()
		upCmd = commands.NewAzureUp(azureClient, boshManager, cloudConfigManager, envIDManager, logger, stateStore, terraformManager)
		deleteLBsCmd = commands.NewAzureDeleteLBs(cloudConfigManager, stateStore, terraformManager)
	}

	awsCreateLBs := commands.NewAWSCreateLBs(cloudConfigManager, stateStore, terraformManager, awsEnvironmentValidator)
	awsUpdateLBs := commands.NewAWSUpdateLBs(awsCreateLBs)

	gcpCreateLBs := commands.NewGCPCreateLBs(terraformManager, cloudConfigManager, stateStore, gcpEnvironmentValidator, gcpClientProvider.Client())
	gcpUpdateLBs := commands.NewGCPUpdateLBs(gcpCreateLBs)

	up := commands.NewUp(upCmd, boshManager)

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
	commandSet["delete-lbs"] = commands.NewDeleteLBs(deleteLBsCmd, logger, stateValidator, boshManager)
	commandSet["lbs"] = commands.NewLBs(lbsCmd, stateValidator)
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

	app := application.New(commandSet, *appConfig, usage)

	err = app.Run()
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}
}
