package deepequal

import (
	"reflect"

	"github.com/pivotal-cf-experimental/gomegamatchers/internal/diff"
)

func Map(expectedMap reflect.Value, actualMap reflect.Value) (bool, diff.Difference) {
	for _, key := range actualMap.MapKeys() {
		if expectedMap.MapIndex(key).Kind() == reflect.Invalid {
			return false, diff.MapExtraKey{
				ExtraKey: key.Interface(),
				AllKeys:  actualMap.MapKeys(),
			}
		}

		equal, difference := Compare(expectedMap.MapIndex(key).Interface(), actualMap.MapIndex(key).Interface())
		if !equal {
			return false, diff.MapNested{
				Key:              key.Interface(),
				NestedDifference: difference,
			}
		}
	}

	for _, key := range expectedMap.MapKeys() {
		if actualMap.MapIndex(key).Kind() == reflect.Invalid {
			return false, diff.MapMissingKey{
				MissingKey: key.Interface(),
				AllKeys:    actualMap.MapKeys(),
			}
		}
	}

	return true, diff.NoDifference{}
}
