package bosh

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/cloudfoundry/bosh-bootloader/fileio"
	"github.com/cloudfoundry/bosh-bootloader/storage"
)

type executorFs interface {
	fileio.FileReader
	fileio.FileWriter
	fileio.Stater
}

type Executor struct {
	CLI             cli
	FS              executorFs
	EmbedData       embed.FS
	EmbedDataPrefix string
}

type DirInput struct {
	StateDir   string
	VarsDir    string
	Deployment string
}

type cli interface {
	GetBOSHPath() string
	Run(stdout io.Writer, workingDirectory string, args []string) error
}

type setupFile struct {
	source   string
	dest     string
	contents []byte
}

const (
	jumpboxDeploymentRepo = "jumpbox-deployment"
	boshDeploymentRepo    = "bosh-deployment"
)

//go:embed deployments/*
var content embed.FS

func NewExecutor(cmd cli, fs executorFs) Executor {
	return Executor{
		CLI:             cmd,
		FS:              fs,
		EmbedData:       content,
		EmbedDataPrefix: "deployments/",
	}
}
func extractNestedFiles(fs embed.FS, fileList []setupFile, path string, trimPrefix string, destPath string, source_entries ...fs.DirEntry) []setupFile {
	for _, source_entry := range source_entries {
		if source_entry.IsDir() {
			dirContents, err := fs.ReadDir(fmt.Sprintf("%v/%v", path, source_entry.Name()))
			if err != nil {
				panic(err)
			}
			fileList = extractNestedFiles(fs, fileList, filepath.Join(path, source_entry.Name()), trimPrefix, destPath, dirContents...)
		} else {
			contents, err := fs.ReadFile(filepath.Join(path, source_entry.Name()))
			if err != nil {
				panic(err)
			}
			fileList = append(fileList, setupFile{
				source:   source_entry.Name(),
				dest:     filepath.Join(destPath, strings.TrimPrefix(path, trimPrefix), source_entry.Name()),
				contents: contents,
			})
		}

	}

	return fileList
}
func (e Executor) getSetupFiles(sourcePath, destPath string) []setupFile {
	files := []setupFile{}
	fullPath := filepath.Join(e.EmbedDataPrefix, sourcePath)

	assetNames, err := e.EmbedData.ReadDir(fullPath)
	prefix := filepath.Join(e.EmbedDataPrefix, sourcePath)

	if err != nil {
		panic(err)
	}
	return extractNestedFiles(e.EmbedData, files, fullPath, prefix, destPath, assetNames...)
}

func (e Executor) PlanJumpbox(input DirInput, deploymentDir, iaas string) error {
	return e.PlanJumpboxWithState(input, deploymentDir, iaas, storage.State{})
}

func (e Executor) PlanJumpboxWithState(input DirInput, deploymentDir, iaas string, state storage.State) error {
	setupFiles := e.getSetupFiles(jumpboxDeploymentRepo, deploymentDir)

	for _, f := range setupFiles {
		// ignore error if dir already exists
		os.MkdirAll(filepath.Dir(f.dest), os.ModePerm) //nolint:errcheck
		err := e.FS.WriteFile(f.dest, f.contents, storage.StateMode)
		if err != nil {
			return fmt.Errorf("jumpbox write setup file: %s", err) //not tested
		}
	}

	sharedArgs := []string{
		"--vars-store", filepath.Join(input.VarsDir, "jumpbox-vars-store.yml"),
		"--vars-file", filepath.Join(input.VarsDir, "jumpbox-vars-file.yml"),
		"-o", filepath.Join(deploymentDir, iaas, "cpi.yml"),
	}

	if iaas == "vsphere" {
		sharedArgs = append(sharedArgs, "-o", filepath.Join(deploymentDir, "vsphere", "resource-pool.yml"))
		vSphereJumpboxNetworkOpsPath := filepath.Join(deploymentDir, "vsphere-jumpbox-network.yml")
		sharedArgs = append(sharedArgs, "-o", vSphereJumpboxNetworkOpsPath)
		err := e.FS.WriteFile(vSphereJumpboxNetworkOpsPath, []byte(VSphereJumpboxNetworkOps), os.ModePerm)
		if err != nil {
			return fmt.Errorf("jumpbox write vsphere network ops file: %s", err) //not tested
		}
	}

	if iaas == "aws" {
		if state.AWS.AssumeRoleArn != "" {
			sharedArgs = append(sharedArgs, "-o", filepath.Join(deploymentDir, "aws", "cpi-assume-role-credentials.yml"))
		}
	}

	jumpboxState := filepath.Join(input.VarsDir, "jumpbox-state.json")

	boshArgs := append([]string{filepath.Join(deploymentDir, "jumpbox.yml"), "--state", jumpboxState}, sharedArgs...)

	switch iaas {
	case "aws":
		boshArgs = append(boshArgs,
			"-v", `access_key_id="${BBL_AWS_ACCESS_KEY_ID}"`,
			"-v", `secret_access_key="${BBL_AWS_SECRET_ACCESS_KEY}"`,
		)
		if state.AWS.AssumeRoleArn != "" {
			boshArgs = append(boshArgs,
				"-v", `role_arn="${BBL_AWS_ASSUME_ROLE}"`,
			)
		}
	case "azure":
		boshArgs = append(boshArgs,
			"-v", `subscription_id="${BBL_AZURE_SUBSCRIPTION_ID}"`,
			"-v", `client_id="${BBL_AZURE_CLIENT_ID}"`,
			"-v", `client_secret="${BBL_AZURE_CLIENT_SECRET}"`,
			"-v", `tenant_id="${BBL_AZURE_TENANT_ID}"`,
		)
	case "gcp":
		boshArgs = append(boshArgs,
			"--var-file", `gcp_credentials_json="${BBL_GCP_SERVICE_ACCOUNT_KEY_PATH}"`,
			"-v", `project_id="${BBL_GCP_PROJECT_ID}"`,
			"-v", `zone="${BBL_GCP_ZONE}"`,
		)
	case "vsphere":
		boshArgs = append(boshArgs,
			"-v", `vcenter_user="${BBL_VSPHERE_VCENTER_USER}"`,
			"-v", `vcenter_password="${BBL_VSPHERE_VCENTER_PASSWORD}"`,
		)
	case "openstack":
		boshArgs = append(boshArgs,
			"-v", `openstack_username="${BBL_OPENSTACK_USERNAME}"`,
			"-v", `openstack_password="${BBL_OPENSTACK_PASSWORD}"`,
		)
	case "cloudstack":
		boshArgs = append(boshArgs,
			"-v", `cloudstack_api_key="${BBL_CLOUDSTACK_API_KEY}"`,
			"-v", `cloudstack_secret_access_key="${BBL_CLOUDSTACK_SECRET_ACCESS_KEY}"`,
		)
	}

	boshPath := e.CLI.GetBOSHPath()

	createEnvCmd := []byte(formatScript(boshPath, input.StateDir, "create-env", boshArgs))
	createJumpboxScript := filepath.Join(input.StateDir, "create-jumpbox.sh")
	err := e.FS.WriteFile(createJumpboxScript, createEnvCmd, 0750)
	if err != nil {
		return err
	}

	deleteEnvCmd := []byte(formatScript(boshPath, input.StateDir, "delete-env", boshArgs))
	deleteJumpboxScript := filepath.Join(input.StateDir, "delete-jumpbox.sh")
	err = e.FS.WriteFile(deleteJumpboxScript, deleteEnvCmd, 0750)
	if err != nil {
		return err
	}

	return nil
}

func (e Executor) getDirectorSetupFiles(stateDir, deploymentDir, iaas string) []setupFile {
	files := e.getSetupFiles(boshDeploymentRepo, deploymentDir)

	statePath := filepath.Join(stateDir, "bbl-ops-files", iaas)
	assetPath := filepath.Join(boshDeploymentRepo, iaas)

	if iaas == "gcp" { //nolint:staticcheck
		files = append(files, setupFile{
			source:   filepath.Join(assetPath, "bosh-director-ephemeral-ip-ops.yml"),
			dest:     filepath.Join(statePath, "bosh-director-ephemeral-ip-ops.yml"),
			contents: []byte(GCPBoshDirectorEphemeralIPOps),
		})
	} else if iaas == "aws" {
		files = append(files, setupFile{
			source:   filepath.Join(assetPath, "bosh-director-ephemeral-ip-ops.yml"),
			dest:     filepath.Join(statePath, "bosh-director-ephemeral-ip-ops.yml"),
			contents: []byte(AWSBoshDirectorEphemeralIPOps),
		})
	}

	return files
}

func (e Executor) getDirectorOpsFiles(stateDir, deploymentDir, iaas string, state storage.State) []string {
	files := []string{
		filepath.Join(deploymentDir, iaas, "cpi.yml"),
		filepath.Join(deploymentDir, "jumpbox-user.yml"),
		filepath.Join(deploymentDir, "uaa.yml"),
		filepath.Join(deploymentDir, "credhub.yml"),
	}
	if iaas == "gcp" { //nolint:staticcheck
		files = append(files, filepath.Join(stateDir, "bbl-ops-files", iaas, "bosh-director-ephemeral-ip-ops.yml"))
	} else if iaas == "aws" {
		files = append(files, filepath.Join(stateDir, "bbl-ops-files", iaas, "bosh-director-ephemeral-ip-ops.yml"))
		files = append(files, filepath.Join(deploymentDir, iaas, "iam-instance-profile.yml"))
		files = append(files, filepath.Join(deploymentDir, iaas, "encrypted-disk.yml"))
		if state.AWS.AssumeRoleArn != "" {
			files = append(files, filepath.Join(deploymentDir, iaas, "cpi-assume-role-credentials.yml"))
		}
	} else if iaas == "vsphere" {
		files = append(files, filepath.Join(deploymentDir, "vsphere", "resource-pool.yml"))
	}
	return files
}

func (e Executor) PlanDirector(input DirInput, deploymentDir, iaas string) error {
	return e.PlanDirectorWithState(input, deploymentDir, iaas, storage.State{})
}

func (e Executor) PlanDirectorWithState(input DirInput, deploymentDir, iaas string, state storage.State) error {
	setupFiles := e.getDirectorSetupFiles(input.StateDir, deploymentDir, iaas)

	for _, f := range setupFiles {
		if f.source != "" {
			os.MkdirAll(filepath.Dir(f.dest), storage.StateMode) //nolint:errcheck
		}
		if err := e.FS.WriteFile(f.dest, f.contents, storage.StateMode); err != nil {
			return fmt.Errorf("director write setup file: %s", err) //not tested
		}
	}

	sharedArgs := []string{
		"--vars-store", filepath.Join(input.VarsDir, "director-vars-store.yml"),
		"--vars-file", filepath.Join(input.VarsDir, "director-vars-file.yml"),
	}

	for _, f := range e.getDirectorOpsFiles(input.StateDir, deploymentDir, iaas, state) {
		sharedArgs = append(sharedArgs, "-o", f)
	}

	boshState := filepath.Join(input.VarsDir, "bosh-state.json")

	boshArgs := append([]string{filepath.Join(deploymentDir, "bosh.yml"), "--state", boshState}, sharedArgs...)

	switch iaas {
	case "aws":
		boshArgs = append(boshArgs,
			"-v", `access_key_id="${BBL_AWS_ACCESS_KEY_ID}"`,
			"-v", `secret_access_key="${BBL_AWS_SECRET_ACCESS_KEY}"`,
		)
		if state.AWS.AssumeRoleArn != "" {
			boshArgs = append(boshArgs,
				"-v", `role_arn="${BBL_AWS_ASSUME_ROLE}"`,
			)
		}
	case "azure":
		boshArgs = append(boshArgs,
			"-v", `subscription_id="${BBL_AZURE_SUBSCRIPTION_ID}"`,
			"-v", `client_id="${BBL_AZURE_CLIENT_ID}"`,
			"-v", `client_secret="${BBL_AZURE_CLIENT_SECRET}"`,
			"-v", `tenant_id="${BBL_AZURE_TENANT_ID}"`,
		)
	case "gcp":
		boshArgs = append(boshArgs,
			"--var-file", `gcp_credentials_json="${BBL_GCP_SERVICE_ACCOUNT_KEY_PATH}"`,
			"-v", `project_id="${BBL_GCP_PROJECT_ID}"`,
			"-v", `zone="${BBL_GCP_ZONE}"`,
		)
	case "vsphere":
		boshArgs = append(boshArgs,
			"-v", `vcenter_user="${BBL_VSPHERE_VCENTER_USER}"`,
			"-v", `vcenter_password="${BBL_VSPHERE_VCENTER_PASSWORD}"`,
		)
	case "openstack":
		boshArgs = append(boshArgs,
			"-v", `openstack_username="${BBL_OPENSTACK_USERNAME}"`,
			"-v", `openstack_password="${BBL_OPENSTACK_PASSWORD}"`,
		)
	case "cloudstack":
		boshArgs = append(boshArgs,
			"-v", `cloudstack_api_key="${BBL_CLOUDSTACK_API_KEY}"`,
			"-v", `cloudstack_secret_access_key="${BBL_CLOUDSTACK_SECRET_ACCESS_KEY}"`,
		)
	}

	boshPath := e.CLI.GetBOSHPath()

	createEnvCmd := []byte(formatScript(boshPath, input.StateDir, "create-env", boshArgs))
	err := e.FS.WriteFile(filepath.Join(input.StateDir, "create-director.sh"), createEnvCmd, 0750)
	if err != nil {
		return err
	}

	deleteEnvCmd := []byte(formatScript(boshPath, input.StateDir, "delete-env", boshArgs))
	err = e.FS.WriteFile(filepath.Join(input.StateDir, "delete-director.sh"), deleteEnvCmd, 0750)
	if err != nil {
		return err
	}

	return nil
}

func formatScript(boshPath, stateDir, command string, args []string) string {
	script := fmt.Sprintf("#!/bin/sh\n%s %s \\\n", boshPath, command)
	for _, arg := range args {
		if arg[0] == '-' {
			script = fmt.Sprintf("%s  %s", script, arg)
		} else {
			script = fmt.Sprintf("%s  %s \\\n", script, arg)
		}
	}
	script = strings.Replace(script, stateDir, "${BBL_STATE_DIR}", -1) //nolint:staticcheck
	return fmt.Sprintf("%s\n", script[:len(script)-2])
}

func (e Executor) WriteDeploymentVars(input DirInput, deploymentVars string) error {
	varsFilePath := filepath.Join(input.VarsDir, fmt.Sprintf("%s-vars-file.yml", input.Deployment))
	err := e.FS.WriteFile(varsFilePath, []byte(deploymentVars), storage.StateMode)
	if err != nil {
		return fmt.Errorf("write vars file: %s", err) // not tested
	}
	return nil
}

func (e Executor) CreateEnv(input DirInput, state storage.State) (string, error) {
	os.Setenv("BBL_STATE_DIR", input.StateDir) //nolint:errcheck
	createEnvScript := filepath.Join(input.StateDir, fmt.Sprintf("create-%s-override.sh", input.Deployment))
	_, err := e.FS.Stat(createEnvScript)
	if err != nil {
		createEnvScript = strings.Replace(createEnvScript, "-override", "", -1) //nolint:staticcheck
	}

	os.Setenv("BBL_IAAS", state.IAAS) //nolint:errcheck
	switch state.IAAS {
	case "aws":
		os.Setenv("BBL_AWS_ACCESS_KEY_ID", state.AWS.AccessKeyID)         //nolint:errcheck
		os.Setenv("BBL_AWS_SECRET_ACCESS_KEY", state.AWS.SecretAccessKey) //nolint:errcheck
	case "azure":
		os.Setenv("BBL_AZURE_CLIENT_ID", state.Azure.ClientID)             //nolint:errcheck
		os.Setenv("BBL_AZURE_CLIENT_SECRET", state.Azure.ClientSecret)     //nolint:errcheck
		os.Setenv("BBL_AZURE_SUBSCRIPTION_ID", state.Azure.SubscriptionID) //nolint:errcheck
		os.Setenv("BBL_AZURE_TENANT_ID", state.Azure.TenantID)             //nolint:errcheck
	case "gcp":
		os.Setenv("BBL_GCP_SERVICE_ACCOUNT_KEY_PATH", state.GCP.ServiceAccountKeyPath) //nolint:errcheck
		os.Setenv("BBL_GCP_ZONE", state.GCP.Zone)                                      //nolint:errcheck
		os.Setenv("BBL_GCP_PROJECT_ID", state.GCP.ProjectID)                           //nolint:errcheck
	case "vsphere":
		os.Setenv("BBL_VSPHERE_VCENTER_USER", state.VSphere.VCenterUser)         //nolint:errcheck
		os.Setenv("BBL_VSPHERE_VCENTER_PASSWORD", state.VSphere.VCenterPassword) //nolint:errcheck
	case "openstack":
		os.Setenv("BBL_OPENSTACK_USERNAME", state.OpenStack.Username) //nolint:errcheck
		os.Setenv("BBL_OPENSTACK_PASSWORD", state.OpenStack.Password) //nolint:errcheck
	case "cloudstack":
		os.Setenv("BBL_CLOUDSTACK_API_KEY", state.CloudStack.ApiKey)                    //nolint:errcheck
		os.Setenv("BBL_CLOUDSTACK_SECRET_ACCESS_KEY", state.CloudStack.SecretAccessKey) //nolint:errcheck
	}

	cmd := exec.Command(createEnvScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return "", fmt.Errorf("running %s: %s", createEnvScript, err)
	}

	name := fmt.Sprintf("%s-vars-store.yml", input.Deployment)
	contents, _ := e.FS.ReadFile(filepath.Join(input.VarsDir, name)) //nolint:errcheck

	return string(contents), nil
}

func (e Executor) DeleteEnv(input DirInput, state storage.State) error {
	isDeletable, err := e.deploymentExists(input.VarsDir, input.Deployment)
	if err != nil {
		return err
	}
	if !isDeletable {
		return nil
	}

	os.Setenv("BBL_STATE_DIR", input.StateDir) //nolint:errcheck

	deleteEnvScript := filepath.Join(input.StateDir, fmt.Sprintf("delete-%s-override.sh", input.Deployment))
	_, err = e.FS.Stat(deleteEnvScript)
	if err != nil {
		deleteEnvScript = strings.Replace(deleteEnvScript, "-override", "", -1) //nolint:staticcheck
	}

	switch state.IAAS {
	case "aws":
		os.Setenv("BBL_AWS_ACCESS_KEY_ID", state.AWS.AccessKeyID)         //nolint:errcheck
		os.Setenv("BBL_AWS_SECRET_ACCESS_KEY", state.AWS.SecretAccessKey) //nolint:errcheck
	case "azure":
		os.Setenv("BBL_AZURE_CLIENT_ID", state.Azure.ClientID)             //nolint:errcheck
		os.Setenv("BBL_AZURE_CLIENT_SECRET", state.Azure.ClientSecret)     //nolint:errcheck
		os.Setenv("BBL_AZURE_SUBSCRIPTION_ID", state.Azure.SubscriptionID) //nolint:errcheck
		os.Setenv("BBL_AZURE_TENANT_ID", state.Azure.TenantID)             //nolint:errcheck
	case "gcp":
		os.Setenv("BBL_GCP_SERVICE_ACCOUNT_KEY_PATH", state.GCP.ServiceAccountKeyPath) //nolint:errcheck
		os.Setenv("BBL_GCP_ZONE", state.GCP.Zone)                                      //nolint:errcheck
		os.Setenv("BBL_GCP_PROJECT_ID", state.GCP.ProjectID)                           //nolint:errcheck
	case "vsphere":
		os.Setenv("BBL_VSPHERE_VCENTER_USER", state.VSphere.VCenterUser)         //nolint:errcheck
		os.Setenv("BBL_VSPHERE_VCENTER_PASSWORD", state.VSphere.VCenterPassword) //nolint:errcheck
	case "cloudstack":
		os.Setenv("BBL_CLOUDSTACK_API_KEY", state.CloudStack.ApiKey)                    //nolint:errcheck
		os.Setenv("BBL_CLOUDSTACK_SECRET_ACCESS_KEY", state.CloudStack.SecretAccessKey) //nolint:errcheck
	}

	cmd := exec.Command(deleteEnvScript)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("Run bosh delete-env %s: %s", input.Deployment, err)
	}

	return nil
}

func (e Executor) deploymentExists(varsDir, deployment string) (bool, error) {
	var deploymentBoshState string
	switch deployment {
	case "director":
		deploymentBoshState = filepath.Join(varsDir, "bosh-state.json")
	case "jumpbox":
		deploymentBoshState = filepath.Join(varsDir, "jumpbox-state.json")
	default:
		return false, fmt.Errorf("Executor doesn't know how to delete a deployed %s", deployment)
	}
	_, err := e.FS.Stat(deploymentBoshState)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (e Executor) Path() string {
	return e.CLI.GetBOSHPath()
}

func (e Executor) Version() (string, error) {
	args := []string{"-v"}
	buffer := bytes.NewBuffer([]byte{})
	err := e.CLI.Run(buffer, "", args)
	if err != nil {
		return "", err
	}

	versionOutput := buffer.String()
	regex := regexp.MustCompile(`\d+.\d+.\d+`)

	version := regex.FindString(versionOutput)
	if version == "" {
		return "", NewBOSHVersionError(errors.New("BOSH version could not be parsed"))
	}

	return version, nil
}
