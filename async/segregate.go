package async

import sh "github.com/daiser/goflow"

func Segregate[V, C comparable](in Flow[V], classificator sh.Classificator[V, C], classes []C) *[]Flow[V] {
	return segregate(in, classificator, classes, false)
}

func SegregateWithUnclassified[V, C comparable](in Flow[V], classificator sh.Classificator[V, C], classes []C) *[]Flow[V] {
	return segregate(in, classificator, classes, true)
}

func segregate[V, C comparable](in Flow[V], classificator sh.Classificator[V, C], classes []C, withUnclassified bool) *[]Flow[V] {
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
	classify     sh.Classificator[V, C]
	outs         []Flow[V]
	routes       map[C]*Flow[V]
	unclassified *Flow[V]
}

func newSegregator[V, C comparable](in Flow[V], classificator sh.Classificator[V, C], classes []C) *segregator[V, C] {
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
