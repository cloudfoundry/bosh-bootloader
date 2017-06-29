package multierror

import (
	"bytes"
	"fmt"
	"strings"
)

type MultiError struct {
	Message string
	Errors  []*MultiError
	isError bool
}

func NewMultiError(message string) *MultiError {
	return &MultiError{
		Message: message,
		Errors:  []*MultiError{},
		isError: false,
	}
}

func (m *MultiError) Add(e error) {
	multierr, ok := e.(*MultiError)
	if ok {
		m.Errors = append(m.Errors, multierr)
	} else {
		leafError := NewMultiError(e.Error())
		leafError.isError = true
		m.Errors = append(m.Errors, leafError)
	}
}

func (m *MultiError) isLeafNode() bool {
	return m.isError && len(m.Errors) == 0
}

func (m *MultiError) Length() int {
	if m.isLeafNode() {
		return 1
	}
	var length int
	for _, err := range m.Errors {
		length += err.Length()
	}
	return length
}

func (m *MultiError) Error() string {
	if m.Length() == 0 {
		return "there were 0 errors"
	}
	return m.formatError(0)
}

func (m *MultiError) getMessage() string {
	if m.isLeafNode() {
		return fmt.Sprintf("* %s", m.Message)
	}
	var grammar string
	if m.Length() == 1 {
		grammar = "was 1 error"
	} else {
		grammar = fmt.Sprintf("were %d errors", m.Length())
	}

	msg := fmt.Sprintf("there %s", grammar)

	if m.Message != "" {
		msg = fmt.Sprintf("%s with '%s'", msg, m.Message)
	}

	return fmt.Sprintf("%s:", msg)
}

func (m *MultiError) formatError(indent int) string {
	var buffer bytes.Buffer
	for i := 0; i < indent; i++ {
		buffer.WriteString("    ")
	}
	whitespace := buffer.String()
	formattedMessage := strings.Replace(m.getMessage(), "\n", fmt.Sprintf("\n%s  ", whitespace), -1)
	buffer.WriteString(fmt.Sprintf("%s\n", formattedMessage))
	for _, elem := range m.Errors {
		buffer.WriteString(elem.formatError(indent + 1))
	}
	return buffer.String()
}
