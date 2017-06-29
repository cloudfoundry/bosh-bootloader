package prettyprint

import (
	"fmt"
	"reflect"
	"strings"
)

func SliceOfValues(values []reflect.Value) string {
	var prettyPrintedValues []string

	for _, value := range values {
		prettyPrintedValues = append(prettyPrintedValues, fmt.Sprintf("<%T> %+v", value.Interface(), value))
	}

	return "[" + strings.Join(prettyPrintedValues, ", ") + "]"
}
