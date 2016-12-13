package fakes

type TerraformExecutor struct {
	ApplyCall struct {
		CallCount int
		Receives  struct {
			Credentials string
			EnvID       string
			ProjectID   string
			Zone        string
			Region      string
			Template    string
			TFState     string
		}
		Returns struct {
			TFState string
			Error   error
		}
	}
	DestroyCall struct {
		CallCount int
		Receives  struct {
			Credentials string
			EnvID       string
			ProjectID   string
			Zone        string
			Region      string
			Template    string
			TFState     string
		}
		Returns struct {
			TFState string
			Error   error
		}
	}
}

func (t *TerraformExecutor) Apply(credentials, envID, projectID, zone, region, template, tfState string) (string, error) {
	t.ApplyCall.CallCount++
	t.ApplyCall.Receives.Credentials = credentials
	t.ApplyCall.Receives.EnvID = envID
	t.ApplyCall.Receives.ProjectID = projectID
	t.ApplyCall.Receives.Zone = zone
	t.ApplyCall.Receives.Region = region
	t.ApplyCall.Receives.Template = template
	t.ApplyCall.Receives.TFState = tfState
	return t.ApplyCall.Returns.TFState, t.ApplyCall.Returns.Error
}

func (t *TerraformExecutor) Destroy(credentials, envID, projectID, zone, region, template, tfState string) (string, error) {
	t.DestroyCall.CallCount++
	t.DestroyCall.Receives.Credentials = credentials
	t.DestroyCall.Receives.EnvID = envID
	t.DestroyCall.Receives.ProjectID = projectID
	t.DestroyCall.Receives.Zone = zone
	t.DestroyCall.Receives.Region = region
	t.DestroyCall.Receives.Template = template
	t.DestroyCall.Receives.TFState = tfState
	return t.DestroyCall.Returns.TFState, t.DestroyCall.Returns.Error
}
