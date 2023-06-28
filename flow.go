package goflow

import (
	"fmt"
	"sync"
)

type Filter[V any] func(V) bool
type Observer[V any] func(V)
type Classificator[V any, C comparable] func(V) []C
type Consumer[V any] func(V)

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

func (f Flow[T]) SendArr(items []T) {
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
