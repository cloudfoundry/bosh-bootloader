package common

type Deletable interface {
	Delete() error
	Name() string
	Type() string
}
