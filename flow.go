package goflow

import (
	"fmt"
	"sync"
)

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

func (in Flow[T]) Filter(filter func(T) bool) Flow[T] {
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

func (in Flow[T]) Segregate(filter func(T) bool) (passed Flow[T], failed Flow[T]) {
	passed = in.new()
	failed = in.new()

	go func() {
		for i := range in.channel {
			if filter(i) {
				passed.channel <- i
			} else {
				failed.channel <- i
			}
		}
		passed.done()
		failed.done()
	}()

	return
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
