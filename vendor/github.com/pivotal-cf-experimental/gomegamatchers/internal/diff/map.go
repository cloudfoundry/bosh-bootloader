package diff

import "reflect"

type MapNested struct {
	Key              interface{}
	NestedDifference Difference
}

type MapExtraKey struct {
	ExtraKey interface{}
	AllKeys  []reflect.Value
}

type MapMissingKey struct {
	MissingKey interface{}
	AllKeys    []reflect.Value
}
