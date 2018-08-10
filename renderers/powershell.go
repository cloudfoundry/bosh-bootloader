package renderers

import (
	"fmt"
	"strings"
)

type powershell struct {
}

// NewPowershell creates a new Powershell Renderer
func NewPowershell() Renderer {
	return &powershell{}
}

func (renderer *powershell) RenderEnvironmentVariable(variable string, value string) string {
	if strings.ContainsAny(value, "\n") {
		suffix := ""
		if !strings.HasSuffix(value, "\n") {
			suffix = "\r\n"
		}
		return fmt.Sprintf("$env:%s='\r\n%s%s'", variable, value, suffix)
	}
	return fmt.Sprintf("$env:%s=\"%s\"", variable, value)
}

func (renderer *powershell) Type() string {
	return ShellTypePowershell
}
