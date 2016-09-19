package fakes

type Usage struct {
	PrintCall struct {
		CallCount int
	}

	PrintCommandUsageCall struct {
		CallCount int
		Receives  struct {
			Command string
			Message string
		}
	}
}

func (u *Usage) Print() {
	u.PrintCall.CallCount++
}

func (u *Usage) PrintCommandUsage(command, message string) {
	u.PrintCommandUsageCall.CallCount++
	u.PrintCommandUsageCall.Receives.Message = message
	u.PrintCommandUsageCall.Receives.Command = command
}
