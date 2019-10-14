package renderers

import (
	"fmt"

	"github.com/cloudfoundry/bosh-bootloader/helpers"
)

type factory struct {
	platform  string
	envGetter helpers.EnvGetter
}

// Factory defines a new renderer factory
type Factory interface {
	Create(shellType string) (Renderer, error)
}

// NewFactory creates a new factory
func NewFactory(envGetter helpers.EnvGetter) Factory {
	return &factory{
		envGetter: envGetter,
	}
}

func (f *factory) createDefault() (Renderer, error) {
	shellType := ShellTypePosix
	value := f.envGetter.Get("PSModulePath")
	if value != "" {
		shellType = ShellTypePowershell
	}
	return f.createFromType(shellType)
}

func (f *factory) createFromType(shellType string) (Renderer, error) {
	switch shellType {
	case ShellTypePowershell:
		return NewPowershell(), nil
	case ShellTypePosix:
		return NewPosix(), nil
	case ShellTypeYaml:
		return NewYaml(), nil
	default:
		return nil, fmt.Errorf("unrecognized type '%s'", shellType)
	}
}

func (f *factory) Create(shellType string) (Renderer, error) {
	if shellType == "" {
		return f.createDefault()
	}
	return f.createFromType(shellType)
}
