package fakes

type FilteredDeleter struct {
	DeleteCall struct {
		CallCount int
		Receives  struct {
			Filter string
		}
		Returns struct {
			Error error
		}
	}

	ListCall struct {
		CallCount int
		Receives  struct {
			Filter string
		}
	}
}

func (l *FilteredDeleter) Delete(filter string) error {
	l.DeleteCall.CallCount++
	l.DeleteCall.Receives.Filter = filter

	return l.DeleteCall.Returns.Error
}

func (l *FilteredDeleter) List(filter string) {
	l.ListCall.CallCount++
	l.ListCall.Receives.Filter = filter
}
