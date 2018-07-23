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
func NewFactory(platform string, envGetter helpers.EnvGetter) Factory {
	return &factory{
		platform:  platform,
		envGetter: envGetter,
	}
}

func (f *factory) createFromPlatform(platform string) (Renderer, error) {
	shellType := ShellTypePosix
	switch platform {
	case "windows":
		value := f.envGetter.Get("CYGWIN")
		if value != "" {
			shellType = ShellTypePosix
		} else {
			shellType = ShellTypePowershell
		}
	default:
		value := f.envGetter.Get("PSModulePath")
		if value != "" {
			shellType = ShellTypePowershell
		} else {
			shellType = ShellTypePosix
		}
	}
	return f.createFromType(shellType)
}

func (f *factory) createFromType(shellType string) (Renderer, error) {
	switch shellType {
	case ShellTypePowershell:
		return NewPowershell(), nil
	case ShellTypePosix:
		return NewPosix(), nil
	default:
		return nil, fmt.Errorf("unrecognized type '%s'", shellType)
	}
}

func (f *factory) Create(shellType string) (Renderer, error) {
	if shellType == "" {
		return f.createFromPlatform(f.platform)
	}
	return f.createFromType(shellType)
}
