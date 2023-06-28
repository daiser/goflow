package async_test

import (
	"fmt"
	"testing"

	"github.com/daiser/goflow/async"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	flow := async.NewFlow[int]()
	evens, _ := flow.Filter(func(i int) bool { return i&1 == 0 }).Collect()
	flow.SendArr([]int{1, 2, 3})
	flow.Wait()

	assert.ElementsMatch(t, *evens, []int{2})
}

func TestSegregate(t *testing.T) {
	flow := async.NewFlow[int]()
	flows := async.Segregate(
		flow,
		func(n int) []string {
			if n&1 == 0 {
				return []string{"even"}
			}
			return []string{"odd"}
		},
		[]string{"even", "odd"},
	)
	evensFlow, oddsFlow := (*flows)[0], (*flows)[1]
	evens, _ := evensFlow.Collect()
	odds, _ := oddsFlow.Collect()

	flow.SendArr([]int{1, 2, 3, 4})
	flow.Wait()

	assert.ElementsMatch(t, *evens, []int{2, 4})
	assert.ElementsMatch(t, *odds, []int{1, 3})
}

func TestSegregateUnclassified(t *testing.T) {
	flow := async.NewFlow[int]()
	flows := async.SegregateWithUnclassified(
		flow,
		func(n int) []string {
			if n&1 == 0 {
				return []string{"even"}
			}
			return []string{"odd"}
		},
		[]string{"even"},
	)
	evensFlow, othersFlow := (*flows)[0], (*flows)[1]
	evens, _ := evensFlow.Collect()
	others, _ := othersFlow.Collect()

	flow.SendArr([]int{1, 2, 3, 4})
	flow.Wait()

	assert.ElementsMatch(t, *evens, []int{2, 4})
	assert.ElementsMatch(t, *others, []int{1, 3})
}

func TestPeep(t *testing.T) {
	flow := async.NewFlow[string]()

	var lastSeen string
	flow.Peep(func(s string) { lastSeen = s }).Dispose()

	flow.SendArr([]string{"one", "two"})
	flow.Wait()

	assert.Equal(t, lastSeen, "two")
}

func TestConsume(t *testing.T) {
	flow := async.NewFlow[string]()

	var lastSeen string
	flow.Consume(func(s string) { lastSeen = s })

	flow.SendArr([]string{"one", "two"})
	flow.Wait()

	assert.Equal(t, lastSeen, "two")
}

func TestMap(t *testing.T) {
	flow := async.NewFlow[int]()

	strings, _ := async.Map(flow, func(n int) string { return fmt.Sprint(n) }).Collect()

	flow.SendArr([]int{1, 2, 3})
	flow.Wait()

	assert.ElementsMatch(t, *strings, []string{"1", "2", "3"})
}
