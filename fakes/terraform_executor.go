package fakes

import "github.com/cloudfoundry/bosh-bootloader/storage"

type Import struct {
	Addr string
	ID   string
}

type TerraformExecutor struct {
	InitCall struct {
		CallCount int
		Receives  struct {
			Template string
			TFState  string
		}
		Returns struct {
			Error error
		}
	}
	ApplyCall struct {
		CallCount int
		Receives  struct {
			Inputs map[string]string
		}
		Returns struct {
			TFState string
			Error   error
		}
	}
	DestroyCall struct {
		CallCount int
		Receives  struct {
			Inputs   map[string]string
			Template string
			TFState  string
		}
		Returns struct {
			TFState string
			Error   error
		}
	}
	ImportCall struct {
		CallCount int
		Receives  struct {
			TFState string
			Imports []Import
			Creds   storage.AWS
		}
		Returns struct {
			TFState string
			Error   error
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
			TFState    string
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
		Receives  struct {
			TFState string
		}
		Returns struct {
			Outputs map[string]interface{}
			Error   error
		}
	}
}

func (t *TerraformExecutor) Init(template, tfState string) error {
	t.InitCall.CallCount++
	t.InitCall.Receives.Template = template
	t.InitCall.Receives.TFState = tfState
	return t.InitCall.Returns.Error
}

func (t *TerraformExecutor) Apply(inputs map[string]string) (string, error) {
	t.ApplyCall.CallCount++
	t.ApplyCall.Receives.Inputs = inputs
	return t.ApplyCall.Returns.TFState, t.ApplyCall.Returns.Error
}

func (t *TerraformExecutor) Destroy(inputs map[string]string, template, tfState string) (string, error) {
	t.DestroyCall.CallCount++
	t.DestroyCall.Receives.Inputs = inputs
	t.DestroyCall.Receives.Template = template
	t.DestroyCall.Receives.TFState = tfState
	return t.DestroyCall.Returns.TFState, t.DestroyCall.Returns.Error
}

func (t *TerraformExecutor) Import(addr, id, tfstate string, creds storage.AWS) (string, error) {
	t.ImportCall.CallCount++
	t.ImportCall.Receives.Imports = append(t.ImportCall.Receives.Imports, Import{
		Addr: addr,
		ID:   id,
	})
	t.ImportCall.Receives.TFState = tfstate
	t.ImportCall.Receives.Creds = creds

	return t.ImportCall.Returns.TFState, t.ImportCall.Returns.Error
}

func (t *TerraformExecutor) Version() (string, error) {
	t.VersionCall.CallCount++
	return t.VersionCall.Returns.Version, t.VersionCall.Returns.Error
}

func (t *TerraformExecutor) Output(tfState, outputName string) (string, error) {
	t.OutputCall.CallCount++
	t.OutputCall.Receives.TFState = tfState
	t.OutputCall.Receives.OutputName = outputName

	if t.OutputCall.Stub != nil {
		return t.OutputCall.Stub(outputName)
	}

	return t.OutputCall.Returns.Output, t.OutputCall.Returns.Error
}

func (t *TerraformExecutor) Outputs(tfState string) (map[string]interface{}, error) {
	t.OutputsCall.CallCount++
	t.OutputsCall.Receives.TFState = tfState

	if t.OutputsCall.Stub != nil {
		return t.OutputsCall.Stub()
	}

	return t.OutputsCall.Returns.Outputs, t.OutputsCall.Returns.Error
}
