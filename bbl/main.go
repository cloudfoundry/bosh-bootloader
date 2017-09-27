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
	"github.com/cloudfoundry/bosh-bootloader/aws/ec2"
	"github.com/cloudfoundry/bosh-bootloader/azure"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/certs"
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/config"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/proxy"
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
	appConfig, err := newConfig.Bootstrap(os.Args)
	log.SetFlags(0)
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
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
	storage.GetStateLogger = stderrLogger
	stateStore := storage.NewStore(appConfig.Global.StateDir)
	stateValidator := application.NewStateValidator(appConfig.Global.StateDir)

	// Terraform
	terraformOutputBuffer := bytes.NewBuffer([]byte{})
	terraformCmd := terraform.NewCmd(os.Stderr, terraformOutputBuffer)
	terraformExecutor := terraform.NewExecutor(terraformCmd, appConfig.Global.Debug)

	var (
		availabilityZoneRetriever ec2.AvailabilityZoneRetriever
		certificateValidator      certs.Validator
		networkClient             helpers.NetworkClient
		networkDeletionValidator  commands.NetworkDeletionValidator

		// this should be replaced by an IAAS agnostic variable, but that needs a common interface. We don't have time right now. AWS clients should also be combined into one struct.
		gcpClient gcp.Client
	)
	if appConfig.State.IAAS == "aws" && needsIAASConfig {
		awsClientProvider := &clientmanager.ClientProvider{}
		awsConfiguration := aws.Config{
			AccessKeyID:     appConfig.State.AWS.AccessKeyID,
			SecretAccessKey: appConfig.State.AWS.SecretAccessKey,
			Region:          appConfig.State.AWS.Region,
		}
		awsClientProvider.SetConfig(awsConfiguration, logger)
		awsClient := awsClientProvider.Client()

		availabilityZoneRetriever = awsClient
		certificateValidator = certs.NewValidator()
		networkDeletionValidator = awsClient

		networkClient = awsClient
	}

	if appConfig.State.IAAS == "gcp" && needsIAASConfig {
		gcpClientProvider := gcp.NewClientProvider(gcpBasePath)
		err = gcpClientProvider.SetConfig(appConfig.State.GCP.ServiceAccountKey, appConfig.State.GCP.ProjectID, appConfig.State.GCP.Region, appConfig.State.GCP.Zone)
		if err != nil {
			log.Fatalf("\n\n%s\n", err)
		}
		gcpClient = gcpClientProvider.Client()
		networkClient = gcpClient
		networkDeletionValidator = gcpClient
	}

	var envIDManager helpers.EnvIDManager
	if appConfig.State.IAAS != "" {
		envIDManager = helpers.NewEnvIDManager(envIDGenerator, networkClient)
	}

	var (
		inputGenerator    terraform.InputGenerator
		outputGenerator   terraform.OutputGenerator
		templateGenerator terraform.TemplateGenerator
	)
	switch appConfig.State.IAAS {
	case "aws":
		templateGenerator = awsterraform.NewTemplateGenerator()
		inputGenerator = awsterraform.NewInputGenerator(availabilityZoneRetriever)
		outputGenerator = awsterraform.NewOutputGenerator(terraformExecutor)
	case "azure":
		templateGenerator = azureterraform.NewTemplateGenerator()
		inputGenerator = azureterraform.NewInputGenerator()
		outputGenerator = azureterraform.NewOutputGenerator(terraformExecutor)
	case "gcp":
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
	})

	// BOSH
	hostKeyGetter := proxy.NewHostKeyGetter()
	socks5Proxy := proxy.NewSocks5Proxy(logger, hostKeyGetter, 0)
	boshCommand := bosh.NewCmd(os.Stderr)
	boshExecutor := bosh.NewExecutor(boshCommand, ioutil.TempDir, ioutil.ReadFile, json.Unmarshal,
		json.Marshal, ioutil.WriteFile)
	boshManager := bosh.NewManager(boshExecutor, logger, socks5Proxy)
	boshClientProvider := bosh.NewClientProvider(socks5Proxy)
	sshKeyGetter := bosh.NewSSHKeyGetter()

	var cloudConfigOpsGenerator cloudconfig.OpsGenerator
	switch appConfig.State.IAAS {
	case "aws":
		cloudConfigOpsGenerator = awscloudconfig.NewOpsGenerator(terraformManager)
	case "gcp":
		cloudConfigOpsGenerator = gcpcloudconfig.NewOpsGenerator(terraformManager)
	case "azure":
		cloudConfigOpsGenerator = azurecloudconfig.NewOpsGenerator(terraformManager)
	}
	cloudConfigManager := cloudconfig.NewManager(logger, boshCommand, cloudConfigOpsGenerator, boshClientProvider, socks5Proxy, terraformManager, sshKeyGetter)

	// Subcommands
	var (
		upCmd        commands.UpCmd
		createLBsCmd commands.CreateLBsCmd
		lbsCmd       commands.LBsCmd
		deleteLBsCmd commands.DeleteLBsCmd
	)
	switch appConfig.State.IAAS {
	case "aws":
		environmentValidator := awsapplication.NewEnvironmentValidator(boshClientProvider)
		upCmd = commands.NewAWSUp()
		createLBsCmd = commands.NewAWSCreateLBs(cloudConfigManager, stateStore, terraformManager, environmentValidator)
		lbsCmd = commands.NewAWSLBs(terraformManager, logger)
		deleteLBsCmd = commands.NewAWSDeleteLBs(cloudConfigManager, stateStore, environmentValidator, terraformManager)
	case "gcp":
		environmentValidator := gcpapplication.NewEnvironmentValidator(boshClientProvider)
		upCmd = commands.NewGCPUp(gcpClient)
		createLBsCmd = commands.NewGCPCreateLBs(terraformManager, cloudConfigManager, stateStore, environmentValidator, gcpClient)
		lbsCmd = commands.NewGCPLBs(terraformManager, logger)
		deleteLBsCmd = commands.NewGCPDeleteLBs(stateStore, environmentValidator, terraformManager, cloudConfigManager)
	case "azure":
		azureClient := azure.NewClient()
		upCmd = commands.NewAzureUp(azureClient)
		deleteLBsCmd = commands.NewAzureDeleteLBs(cloudConfigManager, stateStore, terraformManager)
	}

	// Commands
	up := commands.NewUp(upCmd, boshManager, cloudConfigManager, stateStore, envIDManager, terraformManager)
	usage := commands.NewUsage(logger)

	commandSet := application.CommandSet{}
	commandSet["help"] = usage
	commandSet["version"] = commands.NewVersion(Version, logger)
	commandSet["up"] = up
	sshKeyDeleter := bosh.NewSSHKeyDeleter()
	commandSet["rotate"] = commands.NewRotate(stateValidator, sshKeyDeleter, up)
	commandSet["destroy"] = commands.NewDestroy(logger, os.Stdin, boshManager, stateStore, stateValidator, terraformManager, networkDeletionValidator)
	commandSet["down"] = commandSet["destroy"]
	commandSet["create-lbs"] = commands.NewCreateLBs(createLBsCmd, logger, stateValidator, certificateValidator, boshManager)
	commandSet["update-lbs"] = commandSet["create-lbs"]
	commandSet["delete-lbs"] = commands.NewDeleteLBs(deleteLBsCmd, logger, stateValidator, boshManager)
	commandSet["lbs"] = commands.NewLBs(lbsCmd, stateValidator)
	commandSet["jumpbox-address"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.JumpboxAddressPropertyName)
	commandSet["director-address"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorAddressPropertyName)
	commandSet["director-username"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorUsernamePropertyName)
	commandSet["director-password"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorPasswordPropertyName)
	commandSet["director-ca-cert"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorCACertPropertyName)
	commandSet["ssh-key"] = commands.NewSSHKey(logger, stateValidator, sshKeyGetter)
	commandSet["env-id"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.EnvIDPropertyName)
	commandSet["latest-error"] = commands.NewLatestError(logger, stateValidator)
	commandSet["print-env"] = commands.NewPrintEnv(logger, stateValidator, terraformManager)
	commandSet["cloud-config"] = commands.NewCloudConfig(logger, stateValidator, cloudConfigManager)
	commandSet["jumpbox-deployment-vars"] = commands.NewJumpboxDeploymentVars(logger, boshManager, stateValidator, terraformManager)
	commandSet["bosh-deployment-vars"] = commands.NewBOSHDeploymentVars(logger, boshManager, stateValidator, terraformManager)

	app := application.New(commandSet, appConfig, usage)

	err = app.Run()
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}
}
