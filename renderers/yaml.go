package renderers

import (
	"fmt"
	"strings"
)

type yaml struct {
}

// NewYaml defines a new yaml renderer
func NewYaml() Renderer {
	return &yaml{}
}

func (renderer *yaml) RenderEnvironmentVariable(variable string, value string) string {
	if strings.ContainsAny(value, "\n") {
		value = strings.ReplaceAll(value, "\n", "\\n")
		suffix := ""
		if !strings.HasSuffix(value, "\\n") {
			suffix = "\\n"
		}
		return fmt.Sprintf("%s: \"%s%s\"", strings.ToLower(variable), value, suffix)
	}
	return fmt.Sprintf("%s: \"%s\"", strings.ToLower(variable), value)
}

func (renderer *yaml) Type() string {
	return ShellTypeYaml
}
