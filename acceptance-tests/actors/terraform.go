package actors

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	acceptance "github.com/cloudfoundry/bosh-bootloader/acceptance-tests"
	. "github.com/onsi/ginkgo"
)

const (
	template = `variable "project_id" {
	type = "string"
}

variable "region" {
	type = "string"
}

variable "zone" {
	type = "string"
}

variable "env_id" {
	type = "string"
}

variable "credentials" {
	type = "string"
}

provider "google" {
	credentials = "${file("${var.credentials}")}"
	project = "${var.project_id}"
	region = "${var.region}"
}`
)

type Terraform struct {
	projectID             string
	region                string
	zone                  string
	serviceAccountKeyPath string
}

func NewTerraform(config acceptance.Config) Terraform {
	return Terraform{
		projectID: config.GCPProjectID,
		region:    config.GCPRegion,
		zone:      config.GCPZone,
		serviceAccountKeyPath: config.GCPServiceAccountKeyPath,
	}
}

func (t Terraform) Destroy(state acceptance.State) error {
	tempDir, err := ioutil.TempDir("", "")
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "terraform.tfstate"), []byte(state.TFState()), os.ModePerm)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(filepath.Join(tempDir, "template.tf"), []byte(template), os.ModePerm)
	if err != nil {
		return err
	}

	args := []string{"destroy", "-force"}
	args = append(args, makeVar("project_id", t.projectID)...)
	args = append(args, makeVar("env_id", state.EnvID())...)
	args = append(args, makeVar("region", t.region)...)
	args = append(args, makeVar("zone", t.zone)...)
	args = append(args, makeVar("credentials", t.serviceAccountKeyPath)...)
	runCommand := exec.Command("terraform", args...)
	runCommand.Dir = tempDir
	runCommand.Stdout = GinkgoWriter
	runCommand.Stderr = GinkgoWriter

	return runCommand.Run()
}

func makeVar(name string, value string) []string {
	return []string{"-var", fmt.Sprintf("%s=%s", name, value)}
}
