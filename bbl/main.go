package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
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
	flags "github.com/jessevdk/go-flags"

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
	var global struct {
		Help                 bool   `short:"h"   long:"help"`
		Debug                bool   `short:"d"   long:"debug"`
		Version              bool   `short:"v"   long:"version"`
		StateDir             string `short:"s"   long:"state-dir"`
		IAAS                 string `long:"iaas"                    env:"BBL_IAAS"`
		AWSAccessKeyID       string `long:"aws-access-key-id"       env:"BBL_AWS_ACCESS_KEY_ID"`
		AWSSecretAccessKey   string `long:"aws-secret-access-key"   env:"BBL_AWS_SECRET_ACCESS_KEY"`
		AWSRegion            string `long:"aws-region"              env:"BBL_AWS_REGION"`
		GCPServiceAccountKey string `long:"gcp-service-account-key" env:"BBL_GCP_SERVICE_ACCOUNT_KEY"`
		GCPProjectID         string `long:"gcp-project-id"          env:"BBL_GCP_PROJECT_ID"`
		GCPZone              string `long:"gcp-zone"                env:"BBL_GCP_ZONE"`
		GCPRegion            string `long:"gcp-region"              env:"BBL_GCP_REGION"`
	}

	parser := flags.NewParser(&global, flags.IgnoreUnknown)

	remainingArgs, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		panic(err)
	}

	if global.StateDir == "" {
		var err error
		global.StateDir, err = os.Getwd()
		if err != nil {
			panic(err)
		}
	}

	global.GCPServiceAccountKey, err = parseServiceAccountKey(global.GCPServiceAccountKey)
	if err != nil {
		panic(err)
	}

	loadedState, err := storage.GetState(global.StateDir)
	if err != nil {
		panic(err)
	}

	if global.IAAS != "" {
		loadedState.IAAS = global.IAAS
	}

	if global.AWSRegion != "" {
		loadedState.AWS.Region = global.AWSRegion
	}
	if global.AWSSecretAccessKey != "" {
		loadedState.AWS.SecretAccessKey = global.AWSSecretAccessKey
	}
	if global.AWSAccessKeyID != "" {
		loadedState.AWS.AccessKeyID = global.AWSAccessKeyID
	}

	if global.GCPServiceAccountKey != "" {
		loadedState.GCP.ServiceAccountKey = global.GCPServiceAccountKey
	}
	if global.GCPProjectID != "" {
		loadedState.GCP.ProjectID = global.GCPProjectID
	}
	if global.GCPRegion != "" {
		loadedState.GCP.Region = global.GCPRegion
	}
	if global.GCPZone != "" {
		loadedState.GCP.Zone = global.GCPZone
	}

	// Utilities
	envIDGenerator := helpers.NewEnvIDGenerator(rand.Reader)
	envGetter := helpers.NewEnvGetter()
	logger := application.NewLogger(os.Stdout)
	stderrLogger := application.NewLogger(os.Stderr)

	// Usage Command
	usage := commands.NewUsage(logger)

	storage.GetStateLogger = stderrLogger

	stateStore := storage.NewStore(global.StateDir)
	stateValidator := application.NewStateValidator(global.StateDir)

	awsCredentialValidator := awsapplication.NewCredentialValidator(loadedState.AWS.AccessKeyID, loadedState.AWS.SecretAccessKey, loadedState.AWS.Region)
	gcpCredentialValidator := gcpapplication.NewCredentialValidator(loadedState.GCP.ProjectID, loadedState.GCP.ServiceAccountKey, loadedState.GCP.Region, loadedState.GCP.Zone)
	credentialValidator := application.NewCredentialValidator(loadedState.IAAS, gcpCredentialValidator, awsCredentialValidator)

	// Amazon
	awsConfiguration := aws.Config{
		AccessKeyID:     loadedState.AWS.AccessKeyID,
		SecretAccessKey: loadedState.AWS.SecretAccessKey,
		Region:          loadedState.AWS.Region,
	}

	clientProvider := &clientmanager.ClientProvider{}
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
	if loadedState.IAAS == "gcp" {
		err = gcpClientProvider.SetConfig(loadedState.GCP.ServiceAccountKey, loadedState.GCP.ProjectID, loadedState.GCP.Region, loadedState.GCP.Zone)
		if err != nil {
			panic(err)
		}
	}
	gcpKeyPairUpdater := gcp.NewKeyPairUpdater(rand.Reader, rsa.GenerateKey, ssh.NewPublicKey, gcpClientProvider.Client(), logger)
	gcpKeyPairDeleter := gcp.NewKeyPairDeleter(gcpClientProvider.Client(), logger)
	gcpNetworkInstancesChecker := gcp.NewNetworkInstancesChecker(gcpClientProvider.Client())
	gcpKeyPairManager := gcpkeypair.NewManager(gcpKeyPairUpdater, gcpKeyPairDeleter)

	// EnvID
	envIDManager := helpers.NewEnvIDManager(envIDGenerator, gcpClientProvider.Client(), infrastructureManager)

	// Keypair Manager
	keyPairManager := keypair.NewManager(awsKeyPairManager, gcpKeyPairManager)

	// Terraform
	terraformOutputBuffer := bytes.NewBuffer([]byte{})

	terraformCmd := terraform.NewCmd(os.Stderr, terraformOutputBuffer)
	terraformExecutor := terraform.NewExecutor(terraformCmd, global.Debug)
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
		StateStore:                   stateStore,
		KeyPairManager:               keyPairManager,
		TerraformManager:             terraformManager,
		BoshManager:                  boshManager,
		Logger:                       logger,
		EnvIDManager:                 envIDManager,
		CloudConfigManager:           cloudConfigManager,
		GCPAvailabilityZoneRetriever: gcpClientProvider.Client(),
	})

	gcpCreateLBs := commands.NewGCPCreateLBs(terraformManager, cloudConfigManager, stateStore, logger, gcpClientProvider.Client())

	gcpLBs := commands.NewGCPLBs(terraformManager, logger)

	gcpUpdateLBs := commands.NewGCPUpdateLBs(gcpCreateLBs)

	// Commands
	commandSet := application.CommandSet{}
	commandSet["help"] = usage
	commandSet["version"] = commands.NewVersion(Version, logger)
	commandSet["up"] = commands.NewUp(awsUp, gcpUp, envGetter, boshManager)
	commandSet["destroy"] = commands.NewDestroy(
		credentialValidator, logger, os.Stdin, boshManager, vpcStatusChecker, stackManager,
		infrastructureManager, awsKeyPairDeleter, gcpKeyPairDeleter, certificateDeleter,
		stateStore, stateValidator, terraformManager, gcpNetworkInstancesChecker,
	)
	commandSet["down"] = commandSet["destroy"]
	commandSet["create-lbs"] = commands.NewCreateLBs(awsCreateLBs, gcpCreateLBs, stateValidator, certificateValidator, boshManager)
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
	commandSet["rotate"] = commands.NewRotate(stateStore, keyPairManager, terraformManager, boshManager, stateValidator)

	configuration := &application.Configuration{
		Global: application.GlobalConfiguration{
			StateDir: global.StateDir,
			Debug:    global.Debug,
		},
		State:           loadedState,
		ShowCommandHelp: global.Help,
	}

	if len(remainingArgs) > 0 {
		configuration.Command = remainingArgs[0]
		configuration.SubcommandFlags = remainingArgs[1:]
	} else {
		configuration.ShowCommandHelp = false
		if global.Help {
			configuration.Command = "help"
		}
		if global.Version {
			configuration.Command = "version"
		}
	}

	if len(os.Args) == 1 {
		configuration.Command = "help"
	}

	app := application.New(commandSet, *configuration, usage)

	err = app.Run()
	if err != nil {
		log.Fatalf("\n\n%s\n", err)
	}
}

func parseServiceAccountKey(serviceAccountKey string) (string, error) {
	var key string
	if serviceAccountKey == "" {
		return "", nil
	}

	if _, err := os.Stat(serviceAccountKey); err != nil {
		key = serviceAccountKey
	} else {
		rawServiceAccountKey, err := ioutil.ReadFile(serviceAccountKey)
		if err != nil {
			return "", fmt.Errorf("error reading service account key from file: %v", err)
		}

		key = string(rawServiceAccountKey)
	}

	var tmp interface{}
	err := json.Unmarshal([]byte(key), &tmp)
	if err != nil {
		return "", fmt.Errorf("error unmarshalling service account key (must be valid json): %v", err)
	}

	return key, err
}
