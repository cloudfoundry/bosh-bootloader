package terraform

type OutputGenerator struct {
	executor executor
}

func NewOutputGenerator(executor executor) OutputGenerator {
	return OutputGenerator{
		executor: executor,
	}
}

func (g OutputGenerator) Generate() (Outputs, error) {
	tfOutputs, err := g.executor.Outputs()
	if err != nil {
		return Outputs{}, err
	}

	return Outputs{Map: tfOutputs}, nil
}
