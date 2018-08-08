package renderers

import (
	"fmt"
	"strings"
)

type posix struct {
}

// NewPosix defines a new posix renderer
func NewPosix() Renderer {
	return &posix{}
}

func (renderer *posix) RenderEnvironmentVariable(variable string, value string) string {
	if strings.ContainsAny(value, "\n") {
		suffix := ""
		if !strings.HasSuffix(value, "\n") {
			suffix = "\n"
		}
		return fmt.Sprintf("export %s='%s%s'", variable, value, suffix)
	}
	return fmt.Sprintf("export %s=%s", variable, value)
}

func (renderer *posix) Type() string {
	return ShellTypePosix
}
