package sync_test

import (
	"fmt"
	"testing"

	"github.com/daiser/goflow/sync"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	flow := sync.NewFlow[int]()
	evens := flow.Filter(func(i int) bool { return i&1 == 0 }).Collect()

	flow.Send(1, 2, 3)

	assert.ElementsMatch(t, *evens, []int{2})
}

func TestSegregate(t *testing.T) {
	flow := sync.NewFlow[int]()
	flows := sync.Segregate(
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
	evens := evensFlow.Collect()
	odds := oddsFlow.Collect()

	flow.SendArr([]int{1, 2, 3, 4})

	assert.ElementsMatch(t, *evens, []int{2, 4})
	assert.ElementsMatch(t, *odds, []int{1, 3})
}

func TestPeep(t *testing.T) {
	flow := sync.NewFlow[string]()

	var lastSeen string
	flow.Peep(func(s string) { lastSeen = s })

	flow.SendArr([]string{"one", "two"})

	assert.Equal(t, lastSeen, "two")
}

func TestMap(t *testing.T) {
	flow := sync.NewFlow[int]()

	strings := sync.Map(flow, func(n int) string { return fmt.Sprint(n) }).Collect()

	flow.SendArr([]int{1, 2, 3})

	assert.ElementsMatch(t, *strings, []string{"1", "2", "3"})
}
