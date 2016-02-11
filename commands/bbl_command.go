package commands

import (
	"fmt"
	"io"
	"os"
)

type BBLCommand struct {
	Help                                 func()                                       `short:"h" long:"help" description:"Show usage"`
	Version                              func()                                       `short:"v" long:"version" description:"Show version"`
	AWSConfig                            AWSConfig                                    `group:"aws"`
	UnsupportedPrintConcourseAWSTemplate *UnsupportedPrintConcourseAWSTemplateCommand `command:"unsupported-print-concourse-aws-template"`
	UnsupportedCreateBoshAWSKeypair      UnsupportedCreateBoshAWSKeypairCommand       `command:"unsupported-create-bosh-aws-keypair"`
}

type AWSConfig struct {
	AWSAccessKeyID     string `long:"aws-access-key-id"`
	AWSSecretAccessKey string `long:"aws-secret-access-key"`
	AWSRegion          string `long:"aws-region"`
	EndpointOverride   string `long:"endpoint-override" hidden:"yes"`
}

func NewBBLCommand(stdout io.Writer) *BBLCommand {
	bblCommand := &BBLCommand{
		Version: func() {
			fmt.Fprintln(stdout, "bbl 0.0.1")
			os.Exit(0)
		},
		UnsupportedPrintConcourseAWSTemplate: &UnsupportedPrintConcourseAWSTemplateCommand{
			stdout: stdout,
		},
	}

	bblCommand.UnsupportedCreateBoshAWSKeypair.AWSConfig = &bblCommand.AWSConfig

	return bblCommand
}
