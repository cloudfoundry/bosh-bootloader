package fakes

type StateValidator struct {
	ValidateCall struct {
		CallCount int
		Returns   struct {
			Error error
		}
	}
}

func (s *StateValidator) Validate() error {
	s.ValidateCall.CallCount++
	return s.ValidateCall.Returns.Error
}
