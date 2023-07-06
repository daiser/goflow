package goflow

type Filter[V any] func(V) bool
type Observer[V any] func(V)
type Classify[V any, C comparable] func(V) []C
type Consumer[V any] func(V)
type Mapper[I any, O any] func(I) O
type Selector[I any, O any] func(I) []O

type Optional[V any] struct {
	Value    V
	HasValue bool
}

func Some[V any](value V) Optional[V] {
	return Optional[V]{
		Value:    value,
		HasValue: true,
	}
}

func None[V any]() Optional[V] {
	return Optional[V]{
		HasValue: false,
	}
}
