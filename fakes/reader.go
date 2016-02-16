package fakes

type Reader struct {
	ReadCall struct {
		Returns struct {
			Error error
		}
	}
}

func (r *Reader) Read([]byte) (int, error) {
	return 0, r.ReadCall.Returns.Error
}
