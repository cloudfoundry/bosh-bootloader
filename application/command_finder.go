package application

import "strings"

type CommandFinderResult struct {
	GlobalFlags []string
	Command     string
	OtherArgs   []string
}

type CommandFinder interface {
	FindCommand([]string) CommandFinderResult
}

type commandFinder struct {
}

func NewCommandFinder() CommandFinder {
	return commandFinder{}
}

func (finder commandFinder) FindCommand(input []string) CommandFinderResult {
	commandFinderResult := CommandFinderResult{}

	previousCommand := ""
	commandIndex := 0
	commandFound := false
	for index, word := range input {
		if !strings.HasPrefix(word, "-") {
			if previousCommand != "--state-dir" {
				commandIndex = index
				commandFound = true
				break
			}
		}
		previousCommand = word
	}

	if commandFound {
		commandFinderResult.GlobalFlags = input[:commandIndex]
		commandFinderResult.Command = input[commandIndex]
		commandFinderResult.OtherArgs = input[commandIndex+1:]
	} else {
		commandFinderResult.GlobalFlags = input
	}

	return commandFinderResult
}
