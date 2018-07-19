package renderers

import (
	"bytes"
	"fmt"
	"sort"
	"strings"
)

type bash struct {
}

// NewBash defines a new bash renderer
func NewBash() Renderer {
	return &bash{}
}

func (renderer *bash) RenderEnvironment(environmentVariables map[string]string) string {
	buffer := bytes.Buffer{}

	// sort the keys because the tests will fail if they are out of order
	sorted := make([]string, 0, len(environmentVariables))
	for k := range environmentVariables {
		sorted = append(sorted, k)
	}
	sort.Strings(sorted)

	for _, k := range sorted {
		v := environmentVariables[k]
		buffer.WriteString(renderer.RenderEnvironmentVariable(k, v))
		buffer.WriteRune('\n')
	}
	return buffer.String()
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

func (renderer *bash) RenderProcess(
	path string,
	args []string,
	environmentVariables map[string]string) string {

	result := renderer.RenderEnvironment(environmentVariables)
	result += path
	for _, arg := range args {
		result += " " + arg
	}
	return result + "\n"
}

func (renderer *bash) Shell() string {
	return "bash"
}
