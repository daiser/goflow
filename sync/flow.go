package sync

import (
	"sync"

	"github.com/daiser/goflow"
)

type Flow[V any] struct {
	next      []*Flow[V]
	processor processor[V]
}

func NewFlow[V any]() *Flow[V] {
	return newFlow[V](func(v V) goflow.Optional[V] { return goflow.Some(v) })
}

func newFlow[V any](processor processor[V]) *Flow[V] {
	return &Flow[V]{
		next:      make([]*Flow[V], 0),
		processor: processor,
	}
}

func (f *Flow[V]) Filter(filter goflow.Filter[V]) *Flow[V] {
	return f.attach(createFilter(filter))
}

func (f *Flow[V]) Peep(observer goflow.Observer[V]) *Flow[V] {
	return f.attach(createObserver(observer))
}

func (f *Flow[V]) Collect() *[]V {
	values := make([]V, 0)
	f.attach(func(v V) goflow.Optional[V] {
		values = append(values, v)
		return goflow.None[V]()
	})
	return &values
}

func (f *Flow[V]) SendArr(values []V) {
	for _, value := range values {
		f.accept(value)
	}
}

func (f *Flow[V]) Send(values ...V) {
	f.SendArr(values)
}

func (f *Flow[V]) SendFrom(ch <-chan V) {
	var sending sync.WaitGroup

	sending.Add(1)
	go func() {
		defer sending.Done()
		for value := range ch {
			f.accept(value)
		}
	}()

	sending.Wait()
}

func (f *Flow[V]) attach(processor processor[V]) *Flow[V] {
	subFlow := newFlow(processor)
	f.next = append(f.next, subFlow)
	return subFlow
}

func (f Flow[V]) accept(value V) {
	if result := f.processor(value); result.HasValue {
		for _, route := range f.next {
			route.accept(result.Value)
		}
	}
}

type processor[V any] func(V) goflow.Optional[V]
