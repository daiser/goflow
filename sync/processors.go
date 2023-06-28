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
