package commands

import (
	"fmt"
	"io"
	"os"
)

type BBLCommand struct {
	Help                                 func()                                       `short:"h" long:"help" description:"Show usage"`
	Version                              func()                                       `short:"v" long:"version" description:"Show version"`
	UnsupportedPrintConcourseAWSTemplate *UnsupportedPrintConcourseAWSTemplateCommand `command:"unsupported-print-concourse-aws-template"`
}

func NewBBLCommand(stdout io.Writer) BBLCommand {
	return BBLCommand{
		Version: func() {
			fmt.Fprintln(stdout, "bbl 0.0.1")
			os.Exit(0)
		},
		UnsupportedPrintConcourseAWSTemplate: &UnsupportedPrintConcourseAWSTemplateCommand{
			stdout: stdout,
		},
	}
}
