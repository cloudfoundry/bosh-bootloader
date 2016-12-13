package testhelpers

func Contains(slice []string, word string) bool {
	for _, item := range slice {
		if item == word {
			return true
		}
	}
	return false
}
