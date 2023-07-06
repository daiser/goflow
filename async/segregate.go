package async

import sh "github.com/daiser/goflow"

func Segregate[V, C comparable](in Flow[V], classificator sh.Classify[V, C], classes []C) *[]Flow[V] {
	return segregate(in, classificator, classes, false)
}

func SegregateWithUnclassified[V, C comparable](in Flow[V], classificator sh.Classify[V, C], classes []C) *[]Flow[V] {
	return segregate(in, classificator, classes, true)
}

func segregate[V, C comparable](in Flow[V], classificator sh.Classify[V, C], classes []C, withUnclassified bool) *[]Flow[V] {
	c := sh.NewClassificator(
		classificator,
		classes,
		withUnclassified,
		func() Flow[V] { return in.new() },
	)

	go func() {
		for v := range in.channel {
			routes := c.Find(v)
			for _, flow := range routes {
				flow.channel <- v
			}
		}
		for _, flow := range *c.GetItems() {
			flow.done()
		}
	}()

	return c.GetItems()
}
