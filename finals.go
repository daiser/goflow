package goflow

import "sync"

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

func (in Flow[T]) Consume(consumer Consumer[T]) *sync.WaitGroup {
	var consuming sync.WaitGroup

	consuming.Add(1)
	go func() {
		defer consuming.Done()
		for i := range in.channel {
			consumer(i)
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
