package diff

import "reflect"

type PrimitiveValueMismatch struct {
	ExpectedValue interface{}
	ActualValue   interface{}
}

type PrimitiveTypeMismatch struct {
	ExpectedType reflect.Type
	ActualValue  interface{}
}
