package app

import (
	"fmt"
	"io"
	"strings"
	"sync"
)

type Logger struct {
	newline   bool
	writer    io.Writer
	mutex     *sync.Mutex
	reader    io.Reader
	noConfirm bool
}

// NewLogger returns a new Logger with the provided writer,
// reader, and value of noConfirm.
func NewLogger(writer io.Writer, reader io.Reader, noConfirm bool) *Logger {
	return &Logger{
		newline:   true,
		writer:    writer,
		mutex:     &sync.Mutex{},
		reader:    reader,
		noConfirm: noConfirm,
	}
}

// clear is not threadsafe.
func (l *Logger) clear() {
	if l.newline {
		return
	}

	l.writer.Write([]byte("\n"))
	l.newline = true
}

// Printf handles arguments in the manner of fmt.Fprintf.
func (l *Logger) Printf(message string, a ...interface{}) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.clear()
	fmt.Fprintf(l.writer, "%s", fmt.Sprintf(message, a...))
}

// Println handles the argument in the manner of fmt.Fprintln.
func (l *Logger) Println(message string) {
	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.clear()
	fmt.Fprintln(l.writer, message)
}

// PromptWithDetails will block all other goroutines attempting
// to print the prompt to the logger for a given resource type
// and resource name, while waiting for user input.
func (l *Logger) PromptWithDetails(resourceType, resourceName string) bool {
	if l.noConfirm {
		return true
	}

	l.mutex.Lock()
	defer l.mutex.Unlock()
	l.clear()
	fmt.Fprintf(l.writer, "[%s: %s] Delete? (y/N): ", resourceType, resourceName)
	l.newline = true

	var proceed string
	fmt.Fscanln(l.reader, &proceed)

	proceed = strings.ToLower(proceed)
	if proceed != "yes" && proceed != "y" {
		return false
	}

	return true
}

func (l *Logger) NoConfirm() {
	l.noConfirm = true
}
