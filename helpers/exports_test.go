package helpers

import "regexp"

func SetMatchString(f func(string, string) (bool, error)) {
	matchString = f
}

func ResetMatchString() {
	matchString = regexp.MatchString
}
