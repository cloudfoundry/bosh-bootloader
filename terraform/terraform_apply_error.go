package terraform

type TerraformApplyError struct {
	tfState string
	err     error
}

func NewTerraformApplyError(tfState string, err error) TerraformApplyError {
	return TerraformApplyError{
		tfState: tfState,
		err:     err,
	}
}

func (t TerraformApplyError) Error() string {
	return t.err.Error()
}

func (t TerraformApplyError) TFState() string {
	return t.tfState
}
