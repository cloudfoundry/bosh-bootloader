package renderers

import (
	"fmt"
	"strings"
)

type bash struct {
}

// NewBash defines a new bash renderer
func NewBash() Renderer {
	return &bash{}
}

func (renderer *bash) RenderEnvironmentVariable(variable string, value string) string {
	if strings.ContainsAny(value, "\n") {
		suffix := ""
		if !strings.HasSuffix(value, "\n") {
			suffix = "\n"
		}
		return fmt.Sprintf("export %s='%s%s'", variable, value, suffix)
	}
	return fmt.Sprintf("export %s=%s", variable, value)
}

func (renderer *bash) Shell() string {
	return "bash"
}
