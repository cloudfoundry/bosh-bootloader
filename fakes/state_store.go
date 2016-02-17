package fakes

type StateStore struct {
	MergeCall struct {
		Receives struct {
			Dir string
			Map map[string]interface{}
		}
		Returns struct {
			Error error
		}
	}

	GetStringCall struct {
		Receives struct {
			Dir string
			Key string
		}
		Returns struct {
			Value string
			OK    bool
			Error error
		}
	}
}

func (s *StateStore) Merge(dir string, m map[string]interface{}) error {
	s.MergeCall.Receives.Dir = dir
	s.MergeCall.Receives.Map = m

	return s.MergeCall.Returns.Error
}

func (s *StateStore) GetString(dir, key string) (string, bool, error) {
	s.GetStringCall.Receives.Dir = dir
	s.GetStringCall.Receives.Key = key

	return s.GetStringCall.Returns.Value, s.GetStringCall.Returns.OK, s.GetStringCall.Returns.Error
}
