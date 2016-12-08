package fakes

type AWSKeyPairDeleter struct {
	DeleteCall struct {
		Receives struct {
			Name string
		}
		Returns struct {
			Error error
		}
	}
}

func (d *AWSKeyPairDeleter) Delete(name string) error {
	d.DeleteCall.Receives.Name = name

	return d.DeleteCall.Returns.Error
}
