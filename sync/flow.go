package sync

import "sync"

type Flow[V any] struct {
	next      []*Flow[V]
	processor processor[V]
}

func NewFlow[V any]() *Flow[V] {
	return newFlow[V](func(v V) optional[V] { return some(v) })
}

func newFlow[V any](processor processor[V]) *Flow[V] {
	return &Flow[V]{
		next:      make([]*Flow[V], 0),
		processor: processor,
	}
}

func (f *Flow[V]) Filter(filter Filter[V]) *Flow[V] {
	return f.attach(createFilter(filter))
}

func (f *Flow[V]) Peep(observer Observer[V]) *Flow[V] {
	return f.attach(createObserver(observer))
}

func (f *Flow[V]) Collect() *[]V {
	values := make([]V, 0)
	f.attach(func(v V) optional[V] {
		values = append(values, v)
		return none[V]()
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
	if result := f.processor(value); result.hasValue {
		for _, route := range f.next {
			route.accept(result.value)
		}
	}
}

type processor[V any] func(V) optional[V]

type optional[V any] struct {
	value    V
	hasValue bool
}

func some[V any](value V) optional[V] {
	return optional[V]{
		value:    value,
		hasValue: true,
	}
}

func none[V any]() optional[V] {
	return optional[V]{
		hasValue: false,
	}
}
