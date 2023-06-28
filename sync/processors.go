package sync

type Filter[V any] func(V) bool

func createFilter[V any](filter Filter[V]) processor[V] {
	return func(v V) optional[V] {
		if filter(v) {
			return some(v)
		}
		return none[V]()
	}
}

type Observer[V any] func(V)

func createObserver[V any](observer Observer[V]) processor[V] {
	return func(v V) optional[V] {
		observer(v)
		return some(v)
	}
}

type Mapper[I any, O any] func(I) O

func Map[I any, O any](in *Flow[I], mapper Mapper[I, O]) *Flow[O] {
	out := NewFlow[O]()
	in.attach(createMapper(mapper, out))
	return out
}

func createMapper[I any, O any](mapper Mapper[I, O], out *Flow[O]) processor[I] {
	return func(v I) optional[I] {
		mapped := mapper(v)
		out.accept(mapped)

		return none[I]()
	}
}

type Selector[I any, O any] func(I) []O

func Select[I any, O any](in *Flow[I], selector Selector[I, O]) *Flow[O] {
	out := NewFlow[O]()
	in.attach(createSelector(selector, out))
	return out
}

func createSelector[I any, O any](selector Selector[I, O], out *Flow[O]) processor[I] {
	return func(i I) optional[I] {
		for _, outValue := range selector(i) {
			out.accept(outValue)
		}

		return none[I]()
	}
}

type Classificator[V any, C comparable] func(V) []C

func Segregate[V any, C comparable](in *Flow[V], classify Classificator[V, C], classes []C) *[]*Flow[V] {
	classificator := newClassificator(classify, classes)
	in.attach(createClassificator(classificator))
	return &classificator.outs
}

func createClassificator[V any, C comparable](c classificator[V, C]) processor[V] {
	return func(v V) optional[V] {
		c.accept(v)
		return none[V]()
	}
}

type classificator[V any, C comparable] struct {
	classify Classificator[V, C]
	outs     []*Flow[V]
	routes   map[C]*Flow[V]
}

func newClassificator[V any, C comparable](classify Classificator[V, C], classes []C) classificator[V, C] {
	outs := make([]*Flow[V], len(classes))
	routes := make(map[C]*Flow[V])

	for i, class := range classes {
		outs[i] = NewFlow[V]()
		routes[class] = outs[i]
	}

	return classificator[V, C]{
		classify: classify,
		outs:     outs,
		routes:   routes,
	}
}

func (c classificator[V, C]) accept(v V) {
	classes := c.classify(v)
	for _, class := range classes {
		if route, ok := c.routes[class]; ok {
			route.accept(v)
		}
	}
}

func JoinArr[V any](flows []*Flow[V]) *Flow[V] {
	out := NewFlow[V]()

	for _, flow := range flows {
		flow.attach(func(v V) optional[V] {
			out.accept(v)
			return none[V]()
		})
	}

	return out
}

func Join[V any](flows ...*Flow[V]) *Flow[V] {
	return JoinArr(flows)
}
