package unsupported

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/pivotal-cf-experimental/bosh-bootloader/aws/cloudformation"
	"github.com/pivotal-cf-experimental/bosh-bootloader/commands"
	"github.com/pivotal-cf-experimental/bosh-bootloader/storage"
)

type templateBuilder interface {
	Build() cloudformation.Template
}

type PrintConcourseAWSTemplate struct {
	stdout  io.Writer
	builder templateBuilder
}

func NewPrintConcourseAWSTemplate(stdout io.Writer, builder templateBuilder) PrintConcourseAWSTemplate {
	return PrintConcourseAWSTemplate{
		stdout:  stdout,
		builder: builder,
	}
}

func (c PrintConcourseAWSTemplate) Execute(globalFlags commands.GlobalFlags, state storage.State) (storage.State, error) {
	template := c.builder.Build()
	buf, err := json.MarshalIndent(template, "", "  ")
	if err != nil {
		return storage.State{}, err
	}

	fmt.Fprintf(c.stdout, string(buf))
	return state, nil
}
