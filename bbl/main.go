package main

import (
	"bytes"
	"crypto/rand"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/cloudfoundry/bosh-bootloader/application"
	"github.com/cloudfoundry/bosh-bootloader/aws"
	"github.com/cloudfoundry/bosh-bootloader/azure"
	"github.com/cloudfoundry/bosh-bootloader/backends"
	"github.com/cloudfoundry/bosh-bootloader/bosh"
	"github.com/cloudfoundry/bosh-bootloader/certs"
	"github.com/cloudfoundry/bosh-bootloader/cloudconfig"
	"github.com/cloudfoundry/bosh-bootloader/commands"
	"github.com/cloudfoundry/bosh-bootloader/config"
	"github.com/cloudfoundry/bosh-bootloader/gcp"
	"github.com/cloudfoundry/bosh-bootloader/helpers"
	"github.com/cloudfoundry/bosh-bootloader/renderers"
	"github.com/cloudfoundry/bosh-bootloader/runtimeconfig"
	"github.com/cloudfoundry/bosh-bootloader/ssh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	"github.com/spf13/afero"

	awscloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/aws"
	azurecloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/azure"
	cloudstackcloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/cloudstack"
	gcpcloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	openstackcloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/openstack"
	vspherecloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/vsphere"

	awsterraform "github.com/cloudfoundry/bosh-bootloader/terraform/aws"
	azureterraform "github.com/cloudfoundry/bosh-bootloader/terraform/azure"
	cloudstackterraform "github.com/cloudfoundry/bosh-bootloader/terraform/cloudstack"
	gcpterraform "github.com/cloudfoundry/bosh-bootloader/terraform/gcp"
	openstackterraform "github.com/cloudfoundry/bosh-bootloader/terraform/openstack"
	vsphereterraform "github.com/cloudfoundry/bosh-bootloader/terraform/vsphere"

	awsleftovers "github.com/genevieve/leftovers/aws"
	azureleftovers "github.com/genevieve/leftovers/azure"
	gcpleftovers "github.com/genevieve/leftovers/gcp"
	vsphereleftovers "github.com/genevieve/leftovers/vsphere"
)

var Version = "dev"

func main() {
	log.SetFlags(0)

	logger := application.NewLogger(os.Stdout, os.Stdin)
	stderrLogger := application.NewLogger(os.Stderr, os.Stdin)
	stateBootstrap := storage.NewStateBootstrap(stderrLogger, Version)
	envRendererFactory := renderers.NewFactory(helpers.NewEnvGetter())

	globals, remainingArgs, err := config.ParseArgs(os.Args)
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}
	if globals.NoConfirm {
		logger.NoConfirm()
	}

	// File IO
	fs := afero.NewOsFs()
	afs := &afero.Afero{Fs: fs}

	// bbl Configuration
	garbageCollector := storage.NewGarbageCollector(afs)
	stateStore := storage.NewStore(globals.StateDir, afs, garbageCollector)
	patchDetector := storage.NewPatchDetector(globals.StateDir, logger)
	stateMigrator := storage.NewMigrator(stateStore, afs)
	stateMerger := config.NewMerger(afs)
	storageProvider := backends.NewProvider()
	stateDownloader := config.NewDownloader(storageProvider)
	newConfig := config.NewConfig(stateBootstrap, stateMigrator, stateMerger, stateDownloader, stderrLogger, afs)

	appConfig, err := newConfig.Bootstrap(globals, remainingArgs, len(os.Args))
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}

	// Utilities
	envIDGenerator := helpers.NewEnvIDGenerator(rand.Reader)
	stateValidator := application.NewStateValidator(appConfig.Global.StateDir)
	certificateValidator := certs.NewValidator()
	lbArgsHandler := commands.NewLBArgsHandler(certificateValidator)
	sshCLI := ssh.NewCLI(os.Stdin, os.Stdout, os.Stderr)
	pathFinder := helpers.NewPathFinder()

	// Terraform
	terraformOutputBuffer := bytes.NewBuffer([]byte{})
	dotTerraformDir := filepath.Join(appConfig.Global.StateDir, "terraform", ".terraform")
	bufferingCLI := terraform.NewCLI(terraformOutputBuffer, terraformOutputBuffer, dotTerraformDir, globals.TerraformBinary)
	var (
		terraformCLI terraform.CLI
		out          io.Writer
	)
	if appConfig.Global.Debug {
		errBuffer := io.MultiWriter(os.Stderr, terraformOutputBuffer)
		terraformCLI = terraform.NewCLI(errBuffer, terraformOutputBuffer, dotTerraformDir, globals.TerraformBinary)
		out = os.Stdout
	} else {
		terraformCLI = bufferingCLI
		out = io.Discard
	}
	terraformExecutor := terraform.NewExecutor(terraformCLI, bufferingCLI, stateStore, afs, appConfig.Global.Debug, out)

	// BOSH
	boshPath, err := config.GetBOSHPath()
	if err != nil {
		log.Fatal(err)
	}
	boshCommand := bosh.NewCLI(os.Stderr, boshPath)
	boshExecutor := bosh.NewExecutor(boshCommand, afs)
	sshKeyGetter := bosh.NewSSHKeyGetter(stateStore, afs)
	allProxyGetter := bosh.NewAllProxyGetter(sshKeyGetter, afs)
	credhubGetter := bosh.NewCredhubGetter(stateStore, afs)
	boshCLIProvider := bosh.NewCLIProvider(allProxyGetter, boshPath)
	boshManager := bosh.NewManager(boshExecutor, logger, stateStore, sshKeyGetter, afs, boshCLIProvider)

	configUpdater := bosh.NewConfigUpdater(boshCLIProvider)

	// Clients that require IAAS credentials.
	var (
		// function extract InitializeNetworkClients
		networkClient            helpers.NetworkClient
		networkDeletionValidator commands.NetworkDeletionValidator

		// function extract InitializeLeftovers
		leftovers commands.FilteredDeleter

		awsClient aws.Client
	)
	// IF we could push this whole block down out of main somehow
	if appConfig.CommandModifiesState {
		switch appConfig.State.IAAS {
		case "aws":
			awsClient = aws.NewClient(appConfig.State.AWS, logger)

			networkDeletionValidator = awsClient
			networkClient = awsClient

			sessionToken := ""
			leftovers, err = awsleftovers.NewLeftovers(logger, appConfig.State.AWS.AccessKeyID, appConfig.State.AWS.SecretAccessKey, sessionToken, appConfig.State.AWS.Region)
			if err != nil {
				log.Fatalf("\n\n%s\n", err)
			}

		case "gcp":
			gcpClient, err := gcp.NewClient(appConfig.State.GCP, "")
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

			leftovers, err = gcpleftovers.NewLeftovers(logger, appConfig.State.GCP.ServiceAccountKeyPath)
			if err != nil {
				log.Fatalf("\n\n%s\n", err)
			}

		case "azure":
			azureClient, err := azure.NewClient(appConfig.State.Azure)
			if err != nil {
				log.Fatalf("\n\n%s\n", err)
			}

			networkDeletionValidator = azureClient
			networkClient = azureClient

			leftovers, err = azureleftovers.NewLeftovers(logger, appConfig.State.Azure.ClientID, appConfig.State.Azure.ClientSecret, appConfig.State.Azure.SubscriptionID, appConfig.State.Azure.TenantID)
			if err != nil {
				log.Fatalf("\n\n%s\n", err)
			}
		case "vsphere":
			vSphereLogger := application.NewLogger(os.Stdout, os.Stdin)
			leftovers, err = vsphereleftovers.NewLeftovers(vSphereLogger, appConfig.State.VSphere.VCenterIP, appConfig.State.VSphere.VCenterUser, appConfig.State.VSphere.VCenterPassword, appConfig.State.VSphere.VCenterDC)
			if err != nil {
				log.Fatalf("\n\n%s\n", err)
			}
		}
	}

	// Objects that do not require IAAS credentials.
	var (
		inputGenerator    terraform.InputGenerator
		templateGenerator terraform.TemplateGenerator

		terraformManager        terraform.Manager
		cloudConfigOpsGenerator cloudconfig.OpsGenerator

		lbsCmd commands.LBsCmd
	)
	switch appConfig.State.IAAS {
	case "aws":
		templateGenerator = awsterraform.NewTemplateGenerator()
		inputGenerator = awsterraform.NewInputGenerator(awsClient)

		terraformManager = terraform.NewManager(terraformExecutor, templateGenerator, inputGenerator, terraformOutputBuffer, logger)

		cloudConfigOpsGenerator = awscloudconfig.NewOpsGenerator(terraformManager, awsClient)

		lbsCmd = commands.NewAWSLBs(terraformManager, logger)
	case "azure":
		templateGenerator = azureterraform.NewTemplateGenerator()
		inputGenerator = azureterraform.NewInputGenerator()

		terraformManager = terraform.NewManager(terraformExecutor, templateGenerator, inputGenerator, terraformOutputBuffer, logger)

		cloudConfigOpsGenerator = azurecloudconfig.NewOpsGenerator(terraformManager)

		lbsCmd = commands.NewAzureLBs(terraformManager, logger)
	case "gcp":
		templateGenerator = gcpterraform.NewTemplateGenerator()
		inputGenerator = gcpterraform.NewInputGenerator()

		terraformManager = terraform.NewManager(terraformExecutor, templateGenerator, inputGenerator, terraformOutputBuffer, logger)

		cloudConfigOpsGenerator = gcpcloudconfig.NewOpsGenerator(terraformManager)

		lbsCmd = commands.NewGCPLBs(terraformManager, logger)
	case "vsphere":
		templateGenerator = vsphereterraform.NewTemplateGenerator()
		inputGenerator = vsphereterraform.NewInputGenerator()

		terraformManager = terraform.NewManager(terraformExecutor, templateGenerator, inputGenerator, terraformOutputBuffer, logger)

		cloudConfigOpsGenerator = vspherecloudconfig.NewOpsGenerator(terraformManager)

	case "openstack":
		templateGenerator = openstackterraform.NewTemplateGenerator()
		inputGenerator = openstackterraform.NewInputGenerator()

		terraformManager = terraform.NewManager(terraformExecutor, templateGenerator, inputGenerator, terraformOutputBuffer, logger)

		cloudConfigOpsGenerator = openstackcloudconfig.NewOpsGenerator(terraformManager)
	case "cloudstack":
		templateGenerator = cloudstackterraform.NewTemplateGenerator()
		inputGenerator = cloudstackterraform.NewInputGenerator()

		terraformManager = terraform.NewManager(terraformExecutor, templateGenerator, inputGenerator, terraformOutputBuffer, logger)

		cloudConfigOpsGenerator = cloudstackcloudconfig.NewOpsGenerator(terraformManager)
	}

	cloudConfigManager := cloudconfig.NewManager(logger, configUpdater, stateStore, cloudConfigOpsGenerator, terraformManager, afs)
	runtimeConfigManager := runtimeconfig.NewManager(logger, stateStore, configUpdater, afs)

	// Commands
	var envIDManager helpers.EnvIDManager
	if appConfig.State.IAAS != "" {
		envIDManager = helpers.NewEnvIDManager(envIDGenerator, networkClient)
	}
	plan := commands.NewPlan(boshManager, cloudConfigManager, runtimeConfigManager, stateStore, patchDetector, envIDManager, terraformManager, lbArgsHandler, stderrLogger, Version)
	up := commands.NewUp(plan, boshManager, cloudConfigManager, runtimeConfigManager, stateStore, terraformManager)
	usage := commands.NewUsage(logger)

	commandSet := application.CommandSet{}
	commandSet["help"] = usage
	commandSet["version"] = commands.NewVersion(Version, logger)
	commandSet["outputs"] = commands.NewOutputs(logger, terraformManager, stateValidator)
	commandSet["up"] = up
	commandSet["plan"] = plan
	sshKeyDeleter := bosh.NewSSHKeyDeleter(stateStore, afs)
	commandSet["rotate"] = commands.NewRotate(stateValidator, sshKeyDeleter, up)
	commandSet["destroy"] = commands.NewDestroy(plan, logger, boshManager, stateStore, stateValidator, terraformManager, networkDeletionValidator)
	commandSet["down"] = commandSet["destroy"]
	commandSet["cleanup-leftovers"] = commands.NewCleanupLeftovers(leftovers)
	commandSet["leftovers"] = commandSet["cleanup-leftovers"]
	commandSet["lbs"] = commands.NewLBs(lbsCmd, stateValidator)
	commandSet["jumpbox-address"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.JumpboxAddressPropertyName)
	commandSet["director-address"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorAddressPropertyName)
	commandSet["director-username"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorUsernamePropertyName)
	commandSet["director-password"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorPasswordPropertyName)
	commandSet["director-ca-cert"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.DirectorCACertPropertyName)
	commandSet["ssh-key"] = commands.NewSSHKey(logger, stateValidator, sshKeyGetter)
	commandSet["validate"] = commands.NewValidate(plan, stateStore, terraformManager)
	commandSet["director-ssh-key"] = commands.NewDirectorSSHKey(logger, stateValidator, sshKeyGetter)
	commandSet["env-id"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.EnvIDPropertyName)
	commandSet["latest-error"] = commands.NewLatestError(logger, stateValidator)
	commandSet["print-env"] = commands.NewPrintEnv(logger, stderrLogger, stateValidator, allProxyGetter, credhubGetter, terraformManager, afs, envRendererFactory)
	commandSet["ssh"] = commands.NewSSH(logger, sshCLI, sshKeyGetter, pathFinder, afs, ssh.RandomPort{})

	app := application.New(commandSet, appConfig, usage)

	err = app.Run()
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}
}
