package fakes

import "fmt"

type Logger struct {
	StepCall struct {
		CallCount int
		Receives  struct {
			Message   string
			Arguments []interface{}
		}
		Messages []string
	}

	DotCall struct {
		CallCount int
	}

	PrintfCall struct {
		CallCount int
		Receives  struct {
			Message   string
			Arguments []interface{}
		}
		Messages []string
	}

	PrintlnCall struct {
		CallCount int
		Stub      func(string)
		Receives  struct {
			Message string
		}
		Messages []string
	}

	PromptCall struct {
		CallCount int
		Receives  struct {
			Message string
		}
	}
}

func (l *Logger) Step(message string, a ...interface{}) {
	l.StepCall.CallCount++
	l.StepCall.Receives.Message = message
	l.StepCall.Receives.Arguments = a

	l.StepCall.Messages = append(l.StepCall.Messages, fmt.Sprintf(message, a...))
}

func (l *Logger) Dot() {
	l.DotCall.CallCount++
}

func (l *Logger) Printf(message string, a ...interface{}) {
	l.PrintfCall.CallCount++
	l.PrintfCall.Receives.Message = message
	l.PrintfCall.Receives.Arguments = a

	l.PrintfCall.Messages = append(l.PrintfCall.Messages, fmt.Sprintf(message, a...))
}

func (l *Logger) Println(message string) {
	l.PrintlnCall.CallCount++
	l.PrintlnCall.Receives.Message = message

	if l.PrintlnCall.Stub != nil {
		l.PrintlnCall.Stub(message)
	}

	l.PrintlnCall.Messages = append(l.PrintlnCall.Messages, message)
}

func (l *Logger) Prompt(message string) {
	l.PromptCall.CallCount++
	l.PromptCall.Receives.Message = message
}
