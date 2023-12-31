package async

import "github.com/daiser/goflow"

func Map[I any, O any](in Flow[I], mapper goflow.Mapper[I, O]) Flow[O] {
	out := Flow[O]{
		channel: make(chan O),
		ends:    in.ends,
	}

	go func() {
		for input := range in.channel {
			out.channel <- mapper(input)
		}
		out.done()
	}()

	return out
}
