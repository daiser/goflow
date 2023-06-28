package async

import (
	"errors"
	"sync"
)

func Join[V any](flows []Flow[V]) Flow[V] {
	if len(flows) == 0 {
		panic(errors.New("nothing to join"))
	}

	out := flows[0].new()

	go func() {
		var joining sync.WaitGroup

		for _, flow := range flows {
			joining.Add(1)
			go func(in Flow[V]) {
				defer joining.Done()
				for v := range in.channel {
					out.channel <- v
				}
			}(flow)
		}

		joining.Wait()
		out.done()
	}()

	return out
}

// sugar
func JoinArgs[V any](flows ...Flow[V]) Flow[V] {
	return Join(flows)
}
