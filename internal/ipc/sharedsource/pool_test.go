package sharedsource

import (
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

func TestPoolStoresCompleteBytesBeforePublication(t *testing.T) {
	pool := &Pool{mapping: mappedSources{data: make([]byte, capacity)}}
	parts := []string{"", "\ufeff" + "const café = '😀';\r\n// \x00", string([]byte{0xff})}
	batch, ok := pool.Store(parts)
	if !ok || batch.Length != uint32(len(strings.Join(parts, ""))) {
		t.Fatalf("wrong source length: %+v", batch)
	}
	if got := string(pool.mapping.data[headerSize : headerSize+int(batch.Length)]); got != strings.Join(parts, "") {
		t.Fatalf("source bytes changed: %q", got)
	}
	if got := atomic.LoadUint32((*uint32)(unsafe.Pointer(&pool.mapping.data[0]))); got != batch.Generation {
		t.Fatalf("source not published: %d", got)
	}
	full, ok := pool.Store([]string{strings.Repeat("x", SlotSize)})
	if !ok || full.Length != SlotSize || pool.mapping.data[headerSize+2*SlotSize-1] != 'x' || pool.mapping.data[headerSize+2*SlotSize] != 0 {
		t.Fatal("full slot was truncated or overflowed")
	}
	if _, ok := pool.Store([]string{strings.Repeat("x", SlotSize+1)}); ok {
		t.Fatal("oversized store succeeded")
	}
	if _, ok := pool.Store(nil); ok {
		t.Fatal("absent sources consumed a slot")
	}
}

func TestPoolRejectsStaleReleasesAndBoundsRetention(t *testing.T) {
	pool := &Pool{mapping: mappedSources{data: make([]byte, capacity)}}
	var first Batch
	for i := range slotCount {
		batch, ok := pool.Store([]string{"snapshot"})
		if !ok || batch.Slot != uint32(i) {
			t.Fatalf("slot reused before release: %+v", batch)
		}
		if i == 0 {
			first = batch
		}
	}
	for _, invalid := range []Batch{
		{Slot: slotCount, Generation: first.Generation, Length: first.Length},
		{Slot: first.Slot, Generation: first.Generation + 1, Length: first.Length},
		{Slot: first.Slot, Generation: first.Generation, Length: first.Length + 1},
	} {
		pool.Release(invalid)
		if _, ok := pool.Store([]string{"overwritten"}); ok {
			t.Fatal("invalid release freed an occupied slot")
		}
	}
	pool.Release(first)
	next, ok := pool.Store([]string{"next"})
	if !ok || next.Slot != first.Slot || next.Generation <= first.Generation {
		t.Fatalf("generation did not advance: %+v", next)
	}
	pool.Release(first)
	if _, ok := pool.Store([]string{"overwritten"}); ok {
		t.Fatal("stale release freed a live source")
	}
}

func TestPoolConcurrentReuse(t *testing.T) {
	pool := &Pool{mapping: mappedSources{data: make([]byte, capacity)}}
	var group sync.WaitGroup
	for worker := range 32 {
		group.Go(func() {
			for round := range 100 {
				text := fmt.Sprintf("snapshot-%d-%d-😀", worker, round)
				batch, ok := pool.Store([]string{text})
				if !ok {
					continue
				}
				runtime.Gosched()
				start := headerSize + int(batch.Slot)*SlotSize
				if got := string(pool.mapping.data[start : start+int(batch.Length)]); got != text {
					t.Errorf("source overwritten before release: %q", got)
				}
				pool.Release(batch)
			}
		})
	}
	group.Wait()
}

func TestPoolConcurrentClose(t *testing.T) {
	pool := &Pool{mapping: mappedSources{data: make([]byte, capacity)}}
	closed := 0
	pool.mapping.unmap = func() error {
		// Model invalidating mapped bytes. The race detector must see no
		// simultaneous publication/copy, and subsequent stores must fail.
		clear(pool.mapping.data[:headerSize+16])
		closed++
		return nil
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			for range 100 {
				if batch, ok := pool.Store([]string{"snapshot"}); ok {
					pool.Release(batch)
				}
			}
		})
	}
	runtime.Gosched()
	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
	group.Wait()
	if err := pool.Close(); err != nil {
		t.Fatal(err)
	}
	if _, ok := pool.Store([]string{"after close"}); ok || closed != 1 {
		t.Fatal("closed mapping was used or unmapped twice")
	}
}

func TestPoolCommitFailureDoesNotPublishOrConsumeSlot(t *testing.T) {
	commits := 0
	pool := &Pool{mapping: mappedSources{
		data: make([]byte, headerSize+1),
		commitSlot: func(slot int) error {
			commits++
			if slot != 0 {
				t.Fatalf("failed commit consumed slot zero: %d", slot)
			}
			if commits == 1 {
				return errors.New("commit unavailable")
			}
			return nil
		},
	}}
	if _, ok := pool.Store([]string{"x"}); ok {
		t.Fatal("failed commit accepted source bytes")
	}
	if pool.mapping.data[0] != 0 || pool.mapping.data[headerSize] != 0 {
		t.Fatal("failed commit published or wrote bytes")
	}
	batch, ok := pool.Store([]string{"x"})
	if !ok || batch.Slot != 0 || batch.Generation != 1 {
		t.Fatalf("failed commit consumed a generation: %+v", batch)
	}
	pool.Release(batch)
	if _, ok := pool.Store([]string{"y"}); !ok || commits != 2 {
		t.Fatal("slot backing was not retained across reuse")
	}
}
