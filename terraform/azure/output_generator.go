package azure

type executor interface {
	Outputs(string) (map[string]interface{}, error)
}

type OutputGenerator struct {
	executor executor
}

func NewOutputGenerator(executor executor) OutputGenerator {
	return OutputGenerator{
		executor: executor,
	}
}

func (g OutputGenerator) Generate(tfState string) (map[string]interface{}, error) {
	tfOutputs, err := g.executor.Outputs(tfState)
	if err != nil {
		return map[string]interface{}{}, err
	}

	return tfOutputs, nil
}
