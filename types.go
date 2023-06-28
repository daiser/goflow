package goflow

type Filter[V any] func(V) bool
type Observer[V any] func(V)
type Classificator[V any, C comparable] func(V) []C
type Consumer[V any] func(V)
type Mapper[I any, O any] func(I) O
type Selector[I any, O any] func(I) []O
