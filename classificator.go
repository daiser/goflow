package goflow

type Classificator[V any, S any, C comparable] struct {
	outs         []V
	routes       map[C]V
	unclassified Optional[V]
	classify     Classify[S, C]
}

func NewClassificator[V any, S any, C comparable](
	classify Classify[S, C],
	classes []C,
	withUnclassified bool,
	creator func() V,
) Classificator[V, S, C] {
	outs := make([]V, 0, len(classes)+1)
	routes := make(map[C]V)

	for _, class_ := range classes {
		classV := creator()
		outs = append(outs, classV)
		routes[class_] = classV
	}

	unclassified := None[V]()
	if withUnclassified {
		unclassified = Some(creator())
		outs = append(outs, unclassified.Value)
	}

	return Classificator[V, S, C]{
		outs:         outs,
		routes:       routes,
		unclassified: unclassified,
		classify:     classify,
	}
}

func (c Classificator[V, S, C]) Find(value S) []V {
	vs := make([]V, 0, len(c.outs))

	for _, class_ := range c.classify(value) {
		if v, ok := c.routes[class_]; ok {
			vs = append(vs, v)
		}
	}
	if len(vs) == 0 && c.unclassified.HasValue {
		vs = append(vs, c.unclassified.Value)
	}

	return vs
}

func (c Classificator[V, S, C]) GetItems() *[]V {
	return &c.outs
}
