package renderers

// Renderer defines a rendering interface
type Renderer interface {
	RenderEnvironmentVariable(variable string, value string) string
	Type() string
}
