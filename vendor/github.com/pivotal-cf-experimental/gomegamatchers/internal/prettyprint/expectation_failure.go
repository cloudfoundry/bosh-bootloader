package prettyprint

import (
	"fmt"

	"github.com/pivotal-cf-experimental/gomegamatchers/internal/diff"
)

func ExpectationFailure(difference diff.Difference) string {
	return "error at " + anyFailure(difference)
}

func anyFailure(difference diff.Difference) string {
	switch difference := difference.(type) {
	case diff.NoDifference:
		return ""
	case diff.MapNested:
		return mapNestedFailure(difference)

	case diff.MapMissingKey:
		return mapMissingKeyFailure(difference)

	case diff.MapExtraKey:
		return mapExtraKeyFailure(difference)

	case diff.SliceNested:
		return sliceNestedFailure(difference)

	case diff.SliceExtraElements:
		return sliceExtraElementsFailure(difference)

	case diff.SliceMissingElements:
		return sliceMissingElementsFailure(difference)

	case diff.PrimitiveTypeMismatch:
		return primitiveTypeMismatchFailure(difference)

	case diff.PrimitiveValueMismatch:
		return primitiveValueMismatchFailure(difference)

	default:
		panic("unexpected difference type")
	}
}

func mapNestedFailure(difference diff.MapNested) string {
	return fmt.Sprintf("[%+v]%s", difference.Key, anyFailure(difference.NestedDifference))
}

func mapMissingKeyFailure(difference diff.MapMissingKey) string {
	return fmt.Sprintf(`:
  missing key:
    Expected
        %s
    to contain key
        <%T> %+v`, SliceOfValues(difference.AllKeys), difference.MissingKey, difference.MissingKey)
}

func mapExtraKeyFailure(difference diff.MapExtraKey) string {
	return fmt.Sprintf(`:
  extra key found:
    Expected
        %s
    not to contain key
        <%T> %+v`, SliceOfValues(difference.AllKeys), difference.ExtraKey, difference.ExtraKey)
}

func sliceNestedFailure(difference diff.SliceNested) string {
	return fmt.Sprintf("[%d]%s", difference.Index, anyFailure(difference.NestedDifference))
}

func sliceExtraElementsFailure(difference diff.SliceExtraElements) string {
	return fmt.Sprintf(`:
  extra elements found:
    Expected
        %s
    not to contain elements
        %s`, SliceAsValue(difference.AllElements), SliceAsValue(difference.ExtraElements))
}

func sliceMissingElementsFailure(difference diff.SliceMissingElements) string {
	return fmt.Sprintf(`:
  missing elements:
    Expected
        %s
    to contain elements
        %s`, SliceAsValue(difference.AllElements), SliceAsValue(difference.MissingElements))
}

func primitiveTypeMismatchFailure(difference diff.PrimitiveTypeMismatch) string {
	return fmt.Sprintf(`:
  type mismatch:
    Expected
        <%T> %+v
    to be of type
        <%s>`, difference.ActualValue, difference.ActualValue, difference.ExpectedType)
}

func primitiveValueMismatchFailure(difference diff.PrimitiveValueMismatch) string {
	return fmt.Sprintf(`:
  value mismatch:
    Expected
        <%T> %+v
    to equal
        <%T> %+v`,
		difference.ActualValue, difference.ActualValue,
		difference.ExpectedValue, difference.ExpectedValue)
}
