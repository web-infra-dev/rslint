package utils_test

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/web-infra-dev/rslint/internal/utils"
)

// TestLazyMapBuildsOncePerKey locks in the point of the map: the first caller
// to ask for a key pays for it and every later caller gets what it built.
func TestLazyMapBuildsOncePerKey(t *testing.T) {
	t.Parallel()

	var lazy utils.LazyMap[string, *int]
	builds := 0
	build := func(value int) func() *int {
		return func() *int {
			builds++
			return &value
		}
	}

	first := lazy.Get("a", build(1))
	second := lazy.Get("a", build(2))
	other := lazy.Get("b", build(3))

	if builds != 2 {
		t.Fatalf("build ran %d times, want 2", builds)
	}
	if first != second {
		t.Fatal("the second call for one key returned a different value")
	}
	if *first != 1 || *other != 3 {
		t.Fatalf("got %d and %d, want 1 and 3", *first, *other)
	}
}

// TestLazyMapSharesOneValueUnderRacingBuilds locks in that a race costs a
// redundant build at worst and never a second published value: callers that
// disagree would defeat the point of sharing derived state across a run.
func TestLazyMapSharesOneValueUnderRacingBuilds(t *testing.T) {
	t.Parallel()

	var lazy utils.LazyMap[int, *int]
	var counter atomic.Int64

	const callers = 64
	values := make([]*int, callers)
	start := make(chan struct{})
	var group sync.WaitGroup
	for i := range callers {
		group.Add(1)
		go func() {
			defer group.Done()
			<-start
			values[i] = lazy.Get(0, func() *int {
				value := int(counter.Add(1))
				return &value
			})
		}()
	}
	close(start)
	group.Wait()

	for i, value := range values {
		if value != values[0] {
			t.Fatalf("caller %d got a different value than caller 0", i)
		}
	}
}
