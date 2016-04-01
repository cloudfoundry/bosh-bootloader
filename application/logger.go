package application

import (
	"fmt"
	"io"
)

type Logger struct {
	newline bool
	writer  io.Writer
}

func NewLogger(writer io.Writer) *Logger {
	return &Logger{
		newline: true,
		writer:  writer,
	}
}

func (l *Logger) clear() {
	if l.newline {
		return
	}

	l.writer.Write([]byte("\n"))
	l.newline = true
}

func (l *Logger) Step(message string) {
	l.clear()
	fmt.Fprintf(l.writer, "step: %s\n", message)
	l.newline = true
}

func (l *Logger) Dot() {
	l.writer.Write([]byte("\u2022"))
	l.newline = false
}

func (l *Logger) Println(message string) {
	l.clear()
	fmt.Fprintf(l.writer, "%s\n", message)
}

func (l *Logger) Prompt(message string) {
	l.clear()
	fmt.Fprintf(l.writer, "%s (y/N): ", message)
	l.newline = true
}
