package application

import (
	"fmt"
	"io"
	"strings"
)

type Logger struct {
	newline   bool
	writer    io.Writer
	reader    io.Reader
	noConfirm bool
}

func NewLogger(writer io.Writer, reader io.Reader) *Logger {
	return &Logger{
		newline:   true,
		writer:    writer,
		reader:    reader,
		noConfirm: false,
	}
}

func (l *Logger) clear() {
	if l.newline {
		return
	}

	l.writer.Write([]byte("\n")) //nolint:errcheck
	l.newline = true
}

func (l *Logger) Step(message string, a ...interface{}) {
	l.clear()
	fmt.Fprintf(l.writer, "step: %s\n", fmt.Sprintf(message, a...)) //nolint:errcheck
	l.newline = true
}

func (l *Logger) Dot() {
	l.writer.Write([]byte("\u2022")) //nolint:errcheck
	l.newline = false
}

func (l *Logger) Printf(message string, a ...interface{}) {
	l.clear()
	fmt.Fprintf(l.writer, "%s", fmt.Sprintf(message, a...))
}

func (l *Logger) Println(message string) {
	l.clear()
	fmt.Fprintf(l.writer, "%s\n", message) //nolint:errcheck
}

func (l *Logger) Debugf(message string, a ...interface{}) {
	l.clear()
	fmt.Fprintf(l.writer, "%s\n", message)
}

func (l *Logger) Debugln(message string) {
	l.clear()
	fmt.Fprintf(l.writer, "%s\n", message)
}

func (l *Logger) NoConfirm() {
	l.noConfirm = true
}

func (l *Logger) Prompt(message string) bool {
	if l.noConfirm {
		return true
	}

	l.clear()
	fmt.Fprintf(l.writer, "%s (y/N): ", message)
	l.newline = true

	var proceed string
	_, err := fmt.Fscanln(l.reader, &proceed)
	if err != nil {
		return false
	}

	proceed = strings.ToLower(proceed)
	if proceed == "yes" || proceed == "y" {
		return true
	}
	return false
}

func (l *Logger) PromptWithDetails(resourceType, resourceName string) bool {
	return l.Prompt(fmt.Sprintf("[%s: %s] Delete?", resourceType, resourceName))
}
