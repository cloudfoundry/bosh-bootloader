package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"log"
	"os"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/azure"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/certs"
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/config"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	proxy "github.com/cloudfoundry/socks5-proxy"
	"github.com/spf13/afero"

	awscloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/aws"
	azurecloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/azure"
	gcpcloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	vspherecloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/vsphere"
	awsterraform "github.com/cloudfoundry/bosh-bootloader/terraform/aws"
	azureterraform "github.com/cloudfoundry/bosh-bootloader/terraform/azure"
	gcpterraform "github.com/cloudfoundry/bosh-bootloader/terraform/gcp"
	vsphereterraform "github.com/cloudfoundry/bosh-bootloader/terraform/vsphere"
)

var Version = "dev"

func main() {
	logger := application.NewLogger(os.Stdout)
	stderrLogger := application.NewLogger(os.Stderr)
	stateBootstrap := storage.NewStateBootstrap(stderrLogger, Version)

	globals, _, err := config.ParseArgs(os.Args)
	log.SetFlags(0)
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}

	// File IO
	fs := afero.NewOsFs()
	afs := &afero.Afero{Fs: fs}

	stateStore := storage.NewStore(globals.StateDir, afs)
	stateMigrator := storage.NewMigrator(stateStore, afs)
	newConfig := config.NewConfig(stateBootstrap, stateMigrator, stderrLogger, afs)

	appConfig, err := newConfig.Bootstrap(os.Args)
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}

	needsIAASCreds := config.NeedsIAASCreds(appConfig.Command) && !appConfig.ShowCommandHelp
	if needsIAASCreds {
		err = config.ValidateIAAS(appConfig.State)
		if err != nil {
			log.Fatalf("\n\n%s\n", err)
		}
	}

	// Utilities
	envIDGenerator := helpers.NewEnvIDGenerator(rand.Reader)
	stateValidator := application.NewStateValidator(appConfig.Global.StateDir)
	certificateValidator := certs.NewValidator()
	lbArgsHandler := commands.NewLBArgsHandler(certificateValidator)

	// Terraform
	terraformOutputBuffer := bytes.NewBuffer([]byte{})
	terraformCmd := terraform.NewCmd(os.Stderr, terraformOutputBuffer)
	terraformExecutor := terraform.NewExecutor(terraformCmd, stateStore, afs, appConfig.Global.Debug)

	var (
		networkClient            helpers.NetworkClient
		networkDeletionValidator commands.NetworkDeletionValidator

		gcpClient                 gcp.Client
		availabilityZoneRetriever aws.AvailabilityZoneRetriever
	)
	if appConfig.State.IAAS == "aws" && needsIAASCreds {
		awsClient := aws.NewClient(appConfig.State.AWS, logger)

		availabilityZoneRetriever = awsClient
		networkDeletionValidator = awsClient
		networkClient = awsClient
	} else if appConfig.State.IAAS == "gcp" && needsIAASCreds {
		gcpClient, err = gcp.NewClient(appConfig.State.GCP, "")
		if err != nil {
			log.Fatalf("\n\n%s\n", err)
		}

		networkDeletionValidator = gcpClient
		networkClient = gcpClient

		gcpZonerHack := config.NewGCPZonerHack(gcpClient)
		stateWithZones, err := gcpZonerHack.SetZones(appConfig.State)
		if err != nil {
			log.Fatalf("\n\n%s\n", err)
		}
		appConfig.State = stateWithZones
	} else if appConfig.State.IAAS == "azure" && needsIAASCreds {
		azureClient, err := azure.NewClient(appConfig.State.Azure)
		if err != nil {
			log.Fatalf("\n\n%s\n", err)
		}

		networkDeletionValidator = azureClient
		networkClient = azureClient
	}

	var (
		inputGenerator    terraform.InputGenerator
		templateGenerator terraform.TemplateGenerator
	)
	switch appConfig.State.IAAS {
	case "aws":
		templateGenerator = awsterraform.NewTemplateGenerator()
		inputGenerator = awsterraform.NewInputGenerator(availabilityZoneRetriever)
	case "azure":
		templateGenerator = azureterraform.NewTemplateGenerator()
		inputGenerator = azureterraform.NewInputGenerator()
	case "gcp":
		templateGenerator = gcpterraform.NewTemplateGenerator()
		inputGenerator = gcpterraform.NewInputGenerator()
	case "vsphere":
		templateGenerator = vsphereterraform.NewTemplateGenerator()
		inputGenerator = vsphereterraform.NewInputGenerator()
	}

	terraformManager := terraform.NewManager(terraformExecutor, templateGenerator, inputGenerator, terraformOutputBuffer, logger)

	// BOSH
	hostKeyGetter := proxy.NewHostKeyGetter()
	socks5Proxy := proxy.NewSocks5Proxy(hostKeyGetter)
	boshCommand := bosh.NewCmd(os.Stderr)
	boshExecutor := bosh.NewExecutor(boshCommand, afs.ReadFile, json.Unmarshal, json.Marshal, afs.WriteFile)
	sshKeyGetter := bosh.NewSSHKeyGetter(stateStore, afs)
	credhubGetter := bosh.NewCredhubGetter(stateStore, afs)
	boshManager := bosh.NewManager(boshExecutor, logger, socks5Proxy, stateStore, sshKeyGetter)
	boshClientProvider := bosh.NewClientProvider(socks5Proxy, sshKeyGetter)

	var cloudConfigOpsGenerator cloudconfig.OpsGenerator
	switch appConfig.State.IAAS {
	case "aws":
		cloudConfigOpsGenerator = awscloudconfig.NewOpsGenerator(terraformManager, availabilityZoneRetriever)
	case "azure":
		cloudConfigOpsGenerator = azurecloudconfig.NewOpsGenerator(terraformManager)
	case "gcp":
		cloudConfigOpsGenerator = gcpcloudconfig.NewOpsGenerator(terraformManager)
	case "vsphere":
		cloudConfigOpsGenerator = vspherecloudconfig.NewOpsGenerator(terraformManager)
	}
	cloudConfigManager := cloudconfig.NewManager(logger, boshCommand, stateStore, cloudConfigOpsGenerator, boshClientProvider, terraformManager, sshKeyGetter, afs)

	// Subcommands
	var lbsCmd commands.LBsCmd

	switch appConfig.State.IAAS {
	case "aws":
		lbsCmd = commands.NewAWSLBs(terraformManager, logger)
	case "gcp":
		lbsCmd = commands.NewGCPLBs(terraformManager, logger)
	case "azure":
		lbsCmd = commands.NewAzureLBs(terraformManager, logger)
	}

	// Commands
	var envIDManager helpers.EnvIDManager
	if appConfig.State.IAAS != "" {
		envIDManager = helpers.NewEnvIDManager(envIDGenerator, networkClient)
	}
	plan := commands.NewPlan(boshManager, cloudConfigManager, stateStore, envIDManager, terraformManager, lbArgsHandler, stderrLogger, Version)
	up := commands.NewUp(plan, boshManager, cloudConfigManager, stateStore, terraformManager)
	usage := commands.NewUsage(logger)

	commandSet := application.CommandSet{}
	commandSet["help"] = usage
	commandSet["version"] = commands.NewVersion(Version, logger)
	commandSet["up"] = up
	commandSet["plan"] = plan
	sshKeyDeleter := bosh.NewSSHKeyDeleter(stateStore, afs)
	commandSet["rotate"] = commands.NewRotate(stateValidator, sshKeyDeleter, up)
	commandSet["destroy"] = commands.NewDestroy(plan, logger, os.Stdin, boshManager, stateStore, stateValidator, terraformManager, networkDeletionValidator)
	commandSet["down"] = commandSet["destroy"]
	commandSet["lbs"] = commands.NewLBs(lbsCmd, stateValidator)
	commandSet["jumpbox-address"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.JumpboxAddressPropertyName)
	commandSet["director-address"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorAddressPropertyName)
	commandSet["director-username"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorUsernamePropertyName)
	commandSet["director-password"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorPasswordPropertyName)
	commandSet["director-ca-cert"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorCACertPropertyName)
	commandSet["ssh-key"] = commands.NewSSHKey(logger, stateValidator, sshKeyGetter)
	commandSet["director-ssh-key"] = commands.NewDirectorSSHKey(logger, stateValidator, sshKeyGetter)
	commandSet["env-id"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.EnvIDPropertyName)
	commandSet["latest-error"] = commands.NewLatestError(logger, stateValidator)
	commandSet["print-env"] = commands.NewPrintEnv(logger, stderrLogger, stateValidator, sshKeyGetter, credhubGetter, terraformManager, afs)

	app := application.New(commandSet, appConfig, usage)

	err = app.Run()
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}
}
