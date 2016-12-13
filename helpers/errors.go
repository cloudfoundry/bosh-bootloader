package helpers

import "strings"

type Errors struct {
	errors []string
	length int
}

func NewErrors(args ...string) Errors {
	return Errors{
		errors: append([]string{}, args...),
	}
}

func (e Errors) Error() string {
	if len(e.errors) == 1 {
		return e.errors[0]
	} else {
		errorsList := strings.Join(e.errors, ",\n")
		return "the following errors occurred:\n" + errorsList
	}
}

func (e *Errors) Add(err error) {
	e.errors = append(e.errors, err.Error())
}
