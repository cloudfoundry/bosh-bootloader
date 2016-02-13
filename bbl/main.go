package main

import (
	"crypto/rand"
	"crypto/rsa"
	"fmt"
	"os"

	"github.com/pivotal-cf-experimental/bosh-bootloader/application"
	"github.com/pivotal-cf-experimental/bosh-bootloader/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands/unsupported"
	"github.com/pivotal-cf-experimental/bosh-bootloader/ec2"
	"golang.org/x/crypto/ssh"
)

func main() {
	uuidGenerator := ec2.NewUUIDGenerator(rand.Reader)
	sessionProvider := ec2.NewSessionProvider()

	templateBuilder := cloudformation.NewTemplateBuilder()
	keypairGenerator := ec2.NewKeypairGenerator(rand.Reader, uuidGenerator.Generate, rsa.GenerateKey, ssh.NewPublicKey)
	keypairUploader := ec2.NewKeypairUploader()

	app := application.New(application.CommandSet{
		"help":    commands.NewUsage(os.Stdout),
		"version": commands.NewVersion(os.Stdout),
		"unsupported-print-concourse-aws-template": unsupported.NewPrintConcourseAWSTemplate(os.Stdout, templateBuilder),
		"unsupported-create-bosh-aws-keypair":      unsupported.NewCreateBoshAWSKeypair(keypairGenerator, keypairUploader, sessionProvider),
	}, commands.NewUsage(os.Stdout).Print)

	err := app.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\n%s\n", err)
		os.Exit(1)
	}
}
