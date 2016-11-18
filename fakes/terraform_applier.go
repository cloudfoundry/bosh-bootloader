package fakes

type TerraformApplier struct {
	ApplyCall struct {
		Returns struct {
			TFState string
			Error   error
		}
		Receives struct {
			Credentials string
			EnvID       string
			ProjectID   string
			Zone        string
			Region      string
			Template    string
		}
	}
}

func (t *TerraformApplier) Apply(credentials, envID, projectID, zone, region, template string) (string, error) {
	t.ApplyCall.Receives.Credentials = credentials
	t.ApplyCall.Receives.EnvID = envID
	t.ApplyCall.Receives.ProjectID = projectID
	t.ApplyCall.Receives.Zone = zone
	t.ApplyCall.Receives.Region = region
	t.ApplyCall.Receives.Template = template
	return t.ApplyCall.Returns.TFState, t.ApplyCall.Returns.Error
}
