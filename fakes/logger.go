package fakes

type Logger struct {
	StepCall struct {
		Receives struct {
			Message string
		}
	}

	DotCall struct {
		CallCount int
	}
}

func (l *Logger) Step(message string) {
	l.StepCall.Receives.Message = message
}

func (l *Logger) Dot() {
	l.DotCall.CallCount++
}
