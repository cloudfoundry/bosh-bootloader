package fakes

type Import struct {
	Addr string
	ID   string
}

type TerraformExecutor struct {
	IsInitializedCall struct {
		CallCount int
		Returns   struct {
			IsInitialized bool
		}
	}
	InitCall struct {
		CallCount int
		Receives  struct {
			Template string
			Inputs   map[string]interface{}
		}
		Returns struct {
			Error error
		}
	}
	ApplyCall struct {
		CallCount int
		Receives  struct {
			Credentials map[string]string
		}
		Returns struct {
			Error error
		}
	}
	DestroyCall struct {
		CallCount int
		Receives  struct {
			Credentials map[string]string
		}
		Returns struct {
			Error error
		}
	}
	VersionCall struct {
		CallCount int
		Returns   struct {
			Version string
			Error   error
		}
	}
	OutputCall struct {
		Stub      func(string) (string, error)
		CallCount int
		Receives  struct {
			OutputName string
		}
		Returns struct {
			Output string
			Error  error
		}
	}
	OutputsCall struct {
		Stub      func() (map[string]interface{}, error)
		CallCount int
		Returns   struct {
			Outputs map[string]interface{}
			Error   error
		}
	}
}

func (t *TerraformExecutor) IsInitialized() bool {
	t.IsInitializedCall.CallCount++
	return t.IsInitializedCall.Returns.IsInitialized
}

func (t *TerraformExecutor) Init(template string, inputs map[string]interface{}) error {
	t.InitCall.CallCount++
	t.InitCall.Receives.Template = template
	t.InitCall.Receives.Inputs = inputs
	return t.InitCall.Returns.Error
}

func (t *TerraformExecutor) Apply(credentials map[string]string) error {
	t.ApplyCall.CallCount++
	t.ApplyCall.Receives.Credentials = credentials
	return t.ApplyCall.Returns.Error
}

func (t *TerraformExecutor) Destroy(credentials map[string]string) error {
	t.DestroyCall.CallCount++
	t.DestroyCall.Receives.Credentials = credentials
	return t.DestroyCall.Returns.Error
}

func (t *TerraformExecutor) Version() (string, error) {
	t.VersionCall.CallCount++
	return t.VersionCall.Returns.Version, t.VersionCall.Returns.Error
}

func (t *TerraformExecutor) Output(outputName string) (string, error) {
	t.OutputCall.CallCount++
	t.OutputCall.Receives.OutputName = outputName

	if t.OutputCall.Stub != nil {
		return t.OutputCall.Stub(outputName)
	}

	return t.OutputCall.Returns.Output, t.OutputCall.Returns.Error
}

func (t *TerraformExecutor) Outputs() (map[string]interface{}, error) {
	t.OutputsCall.CallCount++

	if t.OutputsCall.Stub != nil {
		return t.OutputsCall.Stub()
	}

	return t.OutputsCall.Returns.Outputs, t.OutputsCall.Returns.Error
}
