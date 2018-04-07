package cartographer

type tfstate struct {
	Modules []module `json:"modules"`
}

type module struct {
	Path    []string          `json:"path"`
	Outputs map[string]output `json:"outputs"`
}

type output struct {
	Sensitive bool        `json:"sensitive"`
	Type      string      `json:"type"`
	Value     interface{} `json:"value"`
}
