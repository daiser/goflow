package async

import "sync"

type MappingSelector[I any, O any] func(I) <-chan O
type Selector[V any] MappingSelector[V, V]

type MappingArraySelector[I any, O any] func(I) []O
type ArraySelector[V any] MappingArraySelector[V, V]

func Select[I any, O any](in Flow[I], selector MappingSelector[I, O]) Flow[O] {
	out := createOut[I, O](in)

	go func() {
		var selecting sync.WaitGroup
		for inValue := range in.channel {
			selecting.Add(1)
			selection := selector(inValue)

			go func(selection <-chan O) {
				defer selecting.Done()
				for outValue := range selection {
					out.channel <- outValue
				}
			}(selection)
		}
		selecting.Wait()
		out.done()
	}()

	return out
}

func (f Flow[V]) Select(selector Selector[V]) Flow[V] {
	return Select[V, V](f, MappingSelector[V, V](selector))
}

// todo Is this a good idea?
func SelectArr[I any, O any](in Flow[I], selector MappingArraySelector[I, O]) Flow[O] {
	out := createOut[I, O](in)

	go func() {
		for inValue := range in.channel {
			for _, outValue := range selector(inValue) {
				out.channel <- outValue
			}
		}
		out.done()
	}()

	return out
}

func (f Flow[V]) SelectArr(selector ArraySelector[V]) Flow[V] {
	return SelectArr[V, V](f, MappingArraySelector[V, V](selector))
}

func createOut[I any, O any](in Flow[I]) Flow[O] {
	return Flow[O]{
		channel: make(chan O),
		ends:    in.ends,
	}
}
