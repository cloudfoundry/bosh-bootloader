package renderers

import (
	"fmt"
	"os"
)

type factory struct {
	platform string
}

// Factory defines a new renderer factory
type Factory interface {
	Create(shell string) (Renderer, error)
}

// NewFactory creates a new factory
func NewFactory(platform string) Factory {
	return &factory{
		platform: platform,
	}
}

func (f *factory) createFromPlatform(platform string) (Renderer, error) {
	shell := "bash"
	switch platform {
	case "windows":
		shell = "powershell"
	default:
		if _, ok := os.LookupEnv("PSModulePath"); ok {
			shell = "powershell"
		}
	}
	return f.createFromShell(shell)
}

func (f *factory) createFromShell(shell string) (Renderer, error) {
	switch shell {
	case "powershell":
		return NewPowershell(), nil
	case "bash":
		return NewBash(), nil
	default:
		return nil, fmt.Errorf("unrecognized shell '%s'", shell)
	}
}

func (f *factory) Create(shell string) (Renderer, error) {
	if shell == "" {
		return f.createFromPlatform(f.platform)
	}
	return f.createFromShell(shell)
}
