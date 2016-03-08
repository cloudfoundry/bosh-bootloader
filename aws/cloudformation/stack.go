package cloudformation

type Stack struct {
	Name    string
	Status  string
	Outputs map[string]string
}
