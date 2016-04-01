package fakes

type Logger struct {
	StepCall struct {
		CallCount int
		Receives  struct {
			Message string
		}
	}

	DotCall struct {
		CallCount int
	}

	PrintlnCall struct {
		Stub     func(string)
		Receives struct {
			Message string
		}
	}

	PromptCall struct {
		Receives struct {
			Message string
		}
	}
}

func (l *Logger) Step(message string) {
	l.StepCall.CallCount++
	l.StepCall.Receives.Message = message
}

func (l *Logger) Dot() {
	l.DotCall.CallCount++
}

func (l *Logger) Println(message string) {
	l.PrintlnCall.Receives.Message = message

	if l.PrintlnCall.Stub != nil {
		l.PrintlnCall.Stub(message)
	}
}

func (l *Logger) Prompt(message string) {
	l.PromptCall.Receives.Message = message
}
