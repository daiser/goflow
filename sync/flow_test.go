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

func TestSegregateWithUnclassified(t *testing.T) {
    // cooking
    flow := sync.NewFlow[int]()
    flows := sync.SegregateWithUnclassified(
        flow,
        func(n int) []string {
            classes := make([]string, 0, 10)
            for div := 2; div < 10; div++ {
                if n%div ==0 {
                    classes = append(classes, fmt.Sprintf("div%d", div))
                }
            }
            return classes
        },
        []string{"div8"},
    )
    div8, other := (*flows)[0], (*flows)[1]
    div8s := div8.Collect()
    others := other.Collect()

    // running
    for n:=1;n<=50;n++ {
        flow.Send(n)
    }

    // checking
    assert.ElementsMatch(t, *div8s, []int{8,16,24,32,40,48})
    assert.Len(t, *others, 44)
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
