package renderers

// Renderer defines a rendering interface
type Renderer interface {
	RenderEnvironment(environmentVariables map[string]string) string
	RenderEnvironmentVariable(variable string, value string) string
	Shell() string
}
