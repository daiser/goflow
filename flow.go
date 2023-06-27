package goflow

import (
	"fmt"
	"sync"
)

type Filter[V any] func(V) bool
type Observer[V any] func(V)
type Classificator[V any, C comparable] func(V) []C

type Flow[T any] struct {
	channel chan T
	ends    **[]*sync.WaitGroup
}

func NewFlow[T any]() Flow[T] {
	var endsPtrPtr *[]*sync.WaitGroup
	*endsPtrPtr = make([]*sync.WaitGroup, 0) // initial list
	return Flow[T]{
		channel: make(chan T),
		ends:    &endsPtrPtr,
	}
}

func (f Flow[T]) new() Flow[T] {
	return Flow[T]{
		channel: make(chan T),
		ends:    f.ends,
	}
}

func (f Flow[T]) done() {
	close(f.channel)
}

func (f Flow[T]) addEnd(end *sync.WaitGroup) *sync.WaitGroup {
	newPtr := new([]*sync.WaitGroup)
	*newPtr = append(**f.ends, end)
	*f.ends = newPtr
	return end
}

func (f Flow[T]) Wait() {
	for _, waitGroup := range **f.ends {
		waitGroup.Wait()
	}
}

func (f Flow[T]) RunArr(items []T) {
	for _, item := range items {
		f.channel <- item
	}
	f.done()
}

func (f Flow[T]) RunChan(items <-chan T) {
	for item := range items {
		f.channel <- item
	}
	f.done()
}

func (in Flow[T]) Filter(filter Filter[T]) Flow[T] {
	out := in.new()
	go func() {
		for t := range in.channel {
			if filter(t) {
				out.channel <- t
			}
		}
		out.done()
	}()

	return out
}

func (in Flow[T]) Collect() (**[]T, *sync.WaitGroup) {
	var itemsPtrPtr *[]T

	var collecting sync.WaitGroup
	collecting.Add(1)
	go func() {
		defer collecting.Done()

		items := make([]T, 0)
		for i := range in.channel {
			items = append(items, i)
		}
		itemsPtrPtr = &items
	}()

	return &itemsPtrPtr, in.addEnd(&collecting)
}

func (in Flow[T]) Consume(plug func(T)) *sync.WaitGroup {
	var consuming sync.WaitGroup

	consuming.Add(1)
	go func() {
		defer consuming.Done()
		for i := range in.channel {
			plug(i)
		}
	}()

	return in.addEnd(&consuming)
}

func (in Flow[T]) Dispose() *sync.WaitGroup {
	var disposing sync.WaitGroup

	disposing.Add(1)
	go func() {
		defer disposing.Done()
		for range in.channel {
		}
	}()

	return in.addEnd(&disposing)
}

func (in Flow[T]) Tee(n int) []Flow[T] {
	if n < 1 {
		panic(fmt.Errorf("can't tee chain %d times", n))
	}

	outs := make([]Flow[T], n)
	for i := 0; i < n; i++ {
		outs[i] = in.new()
	}

	go func() {
		for t := range in.channel {
			for i := 0; i < n; i++ {
				outs[i].channel <- t
			}
		}
		for i := 0; i < n; i++ {
			outs[i].done()
		}
	}()

	return outs
}

func (in Flow[T]) Peep(observer func(T)) Flow[T] {
	out := in.new()

	go func() {
		for i := range in.channel {
			out.channel <- i
			observer(i)
		}
		out.done()
	}()

	return out
}

func Map[TSrc any, TDst any](in Flow[TSrc], mapper func(TSrc) TDst) Flow[TDst] {
	out := Flow[TDst]{
		channel: make(chan TDst),
		ends:    in.ends,
	}

	go func() {
		for i := range in.channel {
			out.channel <- mapper(i)
		}
		out.done()
	}()

	return out
}

func Segregate[V, C comparable](in Flow[V], classificator Classificator[V, C], classes []C) *[]Flow[V] {
	return segregate(in, classificator, classes, false)
}

func SegregateWithUnclassified[V, C comparable](in Flow[V], classificator Classificator[V, C], classes []C) *[]Flow[V] {
	return segregate(in, classificator, classes, true)
}

func segregate[V, C comparable](in Flow[V], classificator Classificator[V, C], classes []C, withUnclassified bool) *[]Flow[V] {
	s := newSegregator(in, classificator, classes)
	if withUnclassified {
		s.createUnclassified(in)
	}

	go func() {
		for v := range in.channel {
			routes := s.route(v)
			for _, flow := range routes {
				flow.channel <- v
			}
		}
		for _, flow := range s.outs {
			flow.done()
		}
	}()

	return &s.outs
}

type segregator[V any, C comparable] struct {
	classify     Classificator[V, C]
	outs         []Flow[V]
	routes       map[C]*Flow[V]
	unclassified *Flow[V]
}

func newSegregator[V, C comparable](in Flow[V], classificator Classificator[V, C], classes []C) *segregator[V, C] {
	outs := make([]Flow[V], len(classes), len(classes)+1)
	routes := make(map[C]*Flow[V], len(classes))

	for i, class := range classes {
		outs[i] = in.new()
		routes[class] = &outs[i]
	}

	return &segregator[V, C]{
		classify:     classificator,
		outs:         outs,
		routes:       routes,
		unclassified: nil,
	}
}

func (s *segregator[V, C]) createUnclassified(in Flow[V]) {
	s.outs = append(s.outs, in.new())
	s.unclassified = &s.outs[len(s.outs)-1]
}

func (c segregator[V, C]) route(value V) []*Flow[V] {
	destinations := make([]*Flow[V], 0, len(c.outs))

	classes := c.classify(value)
	for _, class := range classes {
		if out, ok := c.routes[class]; ok {
			destinations = append(destinations, out)
		}
	}

	if len(destinations) == 0 && c.unclassified != nil {
		destinations = append(destinations, c.unclassified)
	}

	return destinations
}
