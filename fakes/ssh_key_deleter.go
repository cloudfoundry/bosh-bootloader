package fakes

type SSHKeyDeleter struct {
	DeleteCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
	}
}

func (s *SSHKeyDeleter) Delete() error {
	s.DeleteCall.CallCount++

	return s.DeleteCall.Returns.Error
}
