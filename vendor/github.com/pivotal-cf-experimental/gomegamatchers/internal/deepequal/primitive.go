package deepequal

import (
	"reflect"

	"github.com/pivotal-cf-experimental/gomegamatchers/internal/diff"
)

func Primitive(expectedPrimitive interface{}, actualPrimitive interface{}) (bool, diff.Difference) {
	if !reflect.DeepEqual(expectedPrimitive, actualPrimitive) {
		return false, diff.PrimitiveValueMismatch{
			ExpectedValue: expectedPrimitive,
			ActualValue:   actualPrimitive,
		}
	}

	return true, diff.NoDifference{}
}
