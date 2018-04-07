package main

import (
	"bytes"
	"crypto/rand"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

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
	"github.com/cloudfoundry/bosh-bootloader/ssh"
	"github.com/cloudfoundry/bosh-bootloader/storage"
	"github.com/cloudfoundry/bosh-bootloader/terraform"
	proxy "github.com/cloudfoundry/socks5-proxy"
	"github.com/spf13/afero"

	awscloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/aws"
	azurecloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/azure"
	gcpcloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/gcp"
	openstackcloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/openstack"
	vspherecloudconfig "github.com/cloudfoundry/bosh-bootloader/cloudconfig/vsphere"

	awsterraform "github.com/cloudfoundry/bosh-bootloader/terraform/aws"
	azureterraform "github.com/cloudfoundry/bosh-bootloader/terraform/azure"
	gcpterraform "github.com/cloudfoundry/bosh-bootloader/terraform/gcp"
	openstackterraform "github.com/cloudfoundry/bosh-bootloader/terraform/openstack"
	vsphereterraform "github.com/cloudfoundry/bosh-bootloader/terraform/vsphere"

	"github.com/genevieve/cartographer"
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

	globals, _, err := config.ParseArgs(os.Args)
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
			log.Fatal(err)
		}
	}

	// Utilities
	envIDGenerator := helpers.NewEnvIDGenerator(rand.Reader)
	stateValidator := application.NewStateValidator(appConfig.Global.StateDir)
	certificateValidator := certs.NewValidator()
	lbArgsHandler := commands.NewLBArgsHandler(certificateValidator)
	sshCmd := ssh.NewCmd(os.Stdin, os.Stdout, os.Stderr)
	carto := cartographer.NewCartographer()

	// Terraform
	terraformOutputBuffer := bytes.NewBuffer([]byte{})
	dotTerraformDir := filepath.Join(appConfig.Global.StateDir, "terraform", ".terraform")
	bufferingCmd := terraform.NewCmd(terraformOutputBuffer, terraformOutputBuffer, dotTerraformDir)
	var (
		terraformCmd terraform.Cmd
		out          io.Writer
	)
	if appConfig.Global.Debug {
		errBuffer := io.MultiWriter(os.Stderr, terraformOutputBuffer)
		terraformCmd = terraform.NewCmd(errBuffer, terraformOutputBuffer, dotTerraformDir)
		out = os.Stdout
	} else {
		terraformCmd = bufferingCmd
		out = ioutil.Discard
	}
	terraformExecutor := terraform.NewExecutor(terraformCmd, bufferingCmd, stateStore, afs, appConfig.Global.Debug, out, carto)

	// BOSH
	hostKey := proxy.NewHostKey()
	socks5Proxy := proxy.NewSocks5Proxy(hostKey, nil)
	boshPath, err := config.GetBOSHPath()
	if err != nil {
		log.Fatal(err)
	}
	boshCommand := bosh.NewCmd(os.Stderr, boshPath)
	boshExecutor := bosh.NewExecutor(boshCommand, afs)
	sshKeyGetter := bosh.NewSSHKeyGetter(stateStore, afs)
	allProxyGetter := bosh.NewAllProxyGetter(sshKeyGetter, afs)
	credhubGetter := bosh.NewCredhubGetter(stateStore, afs)
	boshManager := bosh.NewManager(boshExecutor, logger, stateStore, sshKeyGetter, afs, carto)
	boshClientProvider := bosh.NewClientProvider(socks5Proxy, sshKeyGetter)

	// Clients that require IAAS credentials.
	var (
		networkClient            helpers.NetworkClient
		networkDeletionValidator commands.NetworkDeletionValidator

		availabilityZoneRetriever aws.AvailabilityZoneRetriever
		leftovers                 commands.FilteredDeleter
	)
	if needsIAASCreds {
		switch appConfig.State.IAAS {
		case "aws":
			awsClient := aws.NewClient(appConfig.State.AWS, logger)

			availabilityZoneRetriever = awsClient
			networkDeletionValidator = awsClient
			networkClient = awsClient

			leftovers, err = awsleftovers.NewLeftovers(logger, appConfig.State.AWS.AccessKeyID, appConfig.State.AWS.SecretAccessKey, appConfig.State.AWS.Region)
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
		inputGenerator = awsterraform.NewInputGenerator(availabilityZoneRetriever)

		terraformManager = terraform.NewManager(terraformExecutor, templateGenerator, inputGenerator, terraformOutputBuffer, logger)

		cloudConfigOpsGenerator = awscloudconfig.NewOpsGenerator(terraformManager, availabilityZoneRetriever)

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
	}

	cloudConfigManager := cloudconfig.NewManager(logger, boshCommand, stateStore, cloudConfigOpsGenerator, boshClientProvider, terraformManager, afs)

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
	commandSet["outputs"] = commands.NewOutputs(logger, carto, stateStore, stateValidator)
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
	commandSet["director-ssh-key"] = commands.NewDirectorSSHKey(logger, stateValidator, sshKeyGetter)
	commandSet["env-id"] = commands.NewStateQuery(logger, stateValidator, terraformManager, commands.EnvIDPropertyName)
	commandSet["latest-error"] = commands.NewLatestError(logger, stateValidator)
	commandSet["print-env"] = commands.NewPrintEnv(logger, stderrLogger, stateValidator, allProxyGetter, credhubGetter, terraformManager, afs)
	commandSet["ssh"] = commands.NewSSH(sshCmd, sshKeyGetter, afs, ssh.RandomPort{})

	app := application.New(commandSet, appConfig, usage)

	err = app.Run()
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}
}
