package gomegamatchers

import (
	"fmt"
	"os"

	"github.com/onsi/gomega/types"
)

func BeAnOsIsNotExistError() types.GomegaMatcher {
	return &osIsNotExistErrorMatcher{}
}

type osIsNotExistErrorMatcher struct{}

func (matcher *osIsNotExistErrorMatcher) Match(actual interface{}) (success bool, err error) {
	err, ok := actual.(error)
	if !ok {
		return false, fmt.Errorf("BeAnOsIsNotExistError matcher expects an error, got %#v", actual)
	}
	return os.IsNotExist(err), nil
}

func (matcher *osIsNotExistErrorMatcher) FailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %#v\nto be an os.IsNotExist error", actual)
}

func (matcher *osIsNotExistErrorMatcher) NegatedFailureMessage(actual interface{}) (message string) {
	return fmt.Sprintf("Expected %#v\nnot to be an os.IsNotExist error", actual)
}
