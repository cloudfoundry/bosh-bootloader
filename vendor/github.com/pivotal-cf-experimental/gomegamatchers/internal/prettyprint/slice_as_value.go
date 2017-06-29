package prettyprint

import (
	"fmt"
	"reflect"
	"strings"
)

func SliceAsValue(values reflect.Value) string {
	var prettyPrintedValues []string

	for i := 0; i < values.Len(); i++ {
		prettyPrintedValues = append(prettyPrintedValues,
			fmt.Sprintf("<%T> %+v", values.Index(i).Interface(), values.Index(i)))
	}

	return "[" + strings.Join(prettyPrintedValues, ", ") + "]"
}
