package fakes

type TerraformApplier struct {
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
}

func (t *TerraformApplier) Apply(credentials, envID, projectID, zone, region, template, tfState string) (string, error) {
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
