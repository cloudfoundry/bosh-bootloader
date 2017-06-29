package deepequal

import (
	"reflect"

	"github.com/pivotal-cf-experimental/gomegamatchers/internal/diff"
)

func Slice(expectedSlice reflect.Value, actualSlice reflect.Value) (bool, diff.Difference) {
	for i := 0; i < actualSlice.Len(); i++ {
		if i >= expectedSlice.Len() {
			return false, diff.SliceExtraElements{
				ExtraElements: actualSlice.Slice(i, actualSlice.Len()),
				AllElements:   actualSlice,
			}
		}

		equal, difference := Compare(expectedSlice.Index(i).Interface(), actualSlice.Index(i).Interface())
		if !equal {
			return false, diff.SliceNested{
				Index:            i,
				NestedDifference: difference,
			}
		}
	}

	if expectedSlice.Len() > actualSlice.Len() {
		return false, diff.SliceMissingElements{
			MissingElements: expectedSlice.Slice(actualSlice.Len(), expectedSlice.Len()),
			AllElements:     actualSlice,
		}
	}

	return true, diff.NoDifference{}
}
