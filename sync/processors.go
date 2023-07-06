package sync

import "github.com/daiser/goflow"

func createFilter[V any](filter goflow.Filter[V]) processor[V] {
	return func(v V) goflow.Optional[V] {
		if filter(v) {
			return goflow.Some(v)
		}
		return goflow.None[V]()
	}
}

func createObserver[V any](observer goflow.Observer[V]) processor[V] {
	return func(v V) goflow.Optional[V] {
		observer(v)
		return goflow.Some(v)
	}
}

func Map[I any, O any](in *Flow[I], mapper goflow.Mapper[I, O]) *Flow[O] {
	out := NewFlow[O]()
	in.attach(createMapper(mapper, out))
	return out
}

func createMapper[I any, O any](mapper goflow.Mapper[I, O], out *Flow[O]) processor[I] {
	return func(v I) goflow.Optional[I] {
		mapped := mapper(v)
		out.accept(mapped)

		return goflow.None[I]()
	}
}

func Select[I any, O any](in *Flow[I], selector goflow.Selector[I, O]) *Flow[O] {
	out := NewFlow[O]()
	in.attach(createSelector(selector, out))
	return out
}

func createSelector[I any, O any](selector goflow.Selector[I, O], out *Flow[O]) processor[I] {
	return func(i I) goflow.Optional[I] {
		for _, outValue := range selector(i) {
			out.accept(outValue)
		}

		return goflow.None[I]()
	}
}

func Segregate[V any, C comparable](
	in *Flow[V],
	classify goflow.Classify[V, C],
	classes []C,
) *[]*Flow[V] {
	classificator := goflow.NewClassificator(classify, classes, false, func() *Flow[V] { return NewFlow[V]() })
	in.attach(createClassificator(classificator))
	return classificator.GetItems()
}

func SegregateWithUnclassified[V any, C comparable](
	in *Flow[V],
	classify goflow.Classify[V, C],
	classes []C,
) *[]*Flow[V] {
	classificator := goflow.NewClassificator(classify, classes, true, func() *Flow[V] { return NewFlow[V]() })
	in.attach(createClassificator(classificator))
	return classificator.GetItems()
}

func createClassificator[V any, C comparable](c goflow.Classificator[*Flow[V], V, C]) processor[V] {
	return func(v V) goflow.Optional[V] {
		for _, flow := range c.Find(v) {
			flow.accept(v)
		}
		return goflow.None[V]()
	}
}

func JoinArr[V any](flows []*Flow[V]) *Flow[V] {
	out := NewFlow[V]()

	for _, flow := range flows {
		flow.attach(func(v V) goflow.Optional[V] {
			out.accept(v)
			return goflow.None[V]()
		})
	}

	return out
}

func Join[V any](flows ...*Flow[V]) *Flow[V] {
	return JoinArr(flows)
}
