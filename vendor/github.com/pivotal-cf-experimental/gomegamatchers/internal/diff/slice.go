package diff

import "reflect"

type SliceNested struct {
	Index            int
	NestedDifference Difference
}

type SliceExtraElements struct {
	ExtraElements reflect.Value
	AllElements   reflect.Value
}

type SliceMissingElements struct {
	MissingElements reflect.Value
	AllElements     reflect.Value
}
