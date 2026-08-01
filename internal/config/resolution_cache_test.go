package config

import (
	"sync/atomic"
	"testing"
	"time"
)

func TestPublishOnceCacheWaitsForPublishedInitializer(t *testing.T) {
	var cache publishOnceCache[string, int]
	var calls atomic.Int32
	started := make(chan struct{})
	release := make(chan struct{})
	firstResult := make(chan int, 1)
	secondResult := make(chan int, 1)

	go func() {
		firstResult <- cache.getOrInit("same", func() int {
			calls.Add(1)
			close(started)
			<-release
			return 42
		})
	}()
	<-started

	go func() {
		secondResult <- cache.getOrInit("same", func() int {
			calls.Add(1)
			return -1
		})
	}()

	select {
	case value := <-secondResult:
		t.Fatalf("reader returned %d before the published initializer completed", value)
	case <-time.After(20 * time.Millisecond):
	}

	close(release)
	if got := <-firstResult; got != 42 {
		t.Fatalf("first result = %d, want 42", got)
	}
	if got := <-secondResult; got != 42 {
		t.Fatalf("second result = %d, want the published value 42", got)
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("initializer calls = %d, want 1", got)
	}
}

func TestPublishOnceCacheReplaysInitializerPanic(t *testing.T) {
	var cache publishOnceCache[string, int]
	var calls atomic.Int32
	panicValue := &struct{}{}

	for attempt := range 2 {
		var recovered any
		func() {
			defer func() {
				recovered = recover()
			}()
			cache.getOrInit("same", func() int {
				calls.Add(1)
				panic(panicValue)
			})
		}()
		if recovered != panicValue {
			t.Fatalf("attempt %d recovered %#v, want the original panic value", attempt, recovered)
		}
	}
	if got := calls.Load(); got != 1 {
		t.Fatalf("initializer calls = %d, want 1", got)
	}
}
