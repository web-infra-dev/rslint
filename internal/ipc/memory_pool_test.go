package ipc

import (
	"bytes"
	"errors"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"unsafe"
)

func readMemoryRange(t *testing.T, pool *memoryPool, batches []MemoryBatch, span MemoryRange) []byte {
	t.Helper()
	result := make([]byte, 0, span.Length)
	offset, remaining := int(span.Offset), int(span.Length)
	for _, batch := range batches {
		length := int(batch.Length)
		if offset >= length {
			offset -= length
			continue
		}
		n := min(remaining, length-offset)
		start := MemoryHeaderSize + int(batch.Slot)*MemorySlotSize + offset
		result = append(result, pool.mapping.data[start:start+n]...)
		remaining -= n
		offset = 0
		if remaining == 0 {
			return result
		}
	}
	if remaining != 0 {
		t.Fatalf("range exceeds published bytes: %+v in %+v", span, batches)
	}
	return result
}

func TestPoolStoresCompleteBytesBeforePublication(t *testing.T) {
	pool := &memoryPool{mapping: memoryMapping{data: make([]byte, MemoryCapacity)}}
	parts := []Attachment{Text(""), Text("\ufeff" + "const café = '😀';\r\n// \x00"), Bytes([]byte{0, 0xff, 0x80})}
	batches, ranges := pool.store(parts)
	if len(batches) != 1 || batches[0].Length != uint32(parts[1].length()+parts[2].length()) || ranges[0] != nil {
		t.Fatalf("wrong published lengths: %+v, %+v", batches, ranges)
	}
	for i := 1; i < len(parts); i++ {
		want := parts[i].bytes
		if !parts[i].binary {
			want = []byte(parts[i].text)
		}
		if got := readMemoryRange(t, pool, batches, *ranges[i]); !bytes.Equal(got, want) {
			t.Fatalf("attachment %d changed: %q", i, got)
		}
	}
	if got := atomic.LoadUint32((*uint32)(unsafe.Pointer(&pool.mapping.data[0]))); got != batches[0].Generation {
		t.Fatalf("bytes not published: %d", got)
	}
	full, _ := pool.store([]Attachment{Text(strings.Repeat("x", MemorySlotSize))})
	if len(full) != 1 || full[0].Length != MemorySlotSize || pool.mapping.data[MemoryHeaderSize+2*MemorySlotSize-1] != 'x' || pool.mapping.data[MemoryHeaderSize+2*MemorySlotSize] != 0 {
		t.Fatal("full slot was truncated or overflowed")
	}
	if batches, _ := pool.store([]Attachment{Text(""), Bytes(nil)}); len(batches) != 0 {
		t.Fatal("empty attachments consumed a slot")
	}
}

func TestPoolMultipleSlotsAndCompleteCapacityFallback(t *testing.T) {
	pool := &memoryPool{mapping: memoryMapping{data: make([]byte, MemoryHeaderSize+3*MemorySlotSize)}}
	for i := 3; i < MemorySlotCount; i++ {
		pool.slots[i].busy = true
	}
	// The first two values fit a slot individually and retain contiguous
	// physical storage. The larger binary value must span the remaining slots.
	large := bytes.Repeat([]byte{0xff, 0, 0x80, 7}, MemorySlotSize/4+1)
	values := []Attachment{Text("prefix"), Bytes(bytes.Repeat([]byte{1}, MemorySlotSize)), Bytes(large), Text("tail")}
	batches, ranges := pool.store(values)
	if len(batches) != 3 || batches[0].Length != uint32(len("prefix")) || ranges[1].Offset != uint32(len("prefix")) || ranges[2] != nil || ranges[3] == nil {
		t.Fatalf("complete capacity fallback or logical offsets failed: %+v, %+v", batches, ranges)
	}
	// Only one slot was left, so the >slot binary value stayed completely
	// inline; its failure did not reserve that last slot against a small value.
	if got := readMemoryRange(t, pool, batches, *ranges[3]); string(got) != "tail" {
		t.Fatalf("fallback consumed or changed later attachment: %q", got)
	}
	pool.release(batches)
	batches, ranges = pool.store([]Attachment{Text("prefix"), Bytes(large), Text("tail")})
	if len(batches) != 2 || ranges[1] == nil || ranges[1].Length != uint32(len(large)) || ranges[2] == nil {
		t.Fatalf("large attachment did not span slots in one request: %+v, %+v", batches, ranges)
	}
	if got := readMemoryRange(t, pool, batches, *ranges[1]); !bytes.Equal(got, large) {
		t.Fatal("straddling binary attachment was changed or truncated")
	}
	if got := readMemoryRange(t, pool, batches, *ranges[2]); string(got) != "tail" {
		t.Fatal("attachment after a straddling value was changed")
	}
}

func TestPoolRejectsStaleReleasesAndBoundsRetention(t *testing.T) {
	pool := &memoryPool{mapping: memoryMapping{data: make([]byte, MemoryCapacity)}}
	var first MemoryBatch
	for i := range MemorySlotCount {
		batches, _ := pool.store([]Attachment{Text("snapshot")})
		if len(batches) != 1 || batches[0].Slot != uint32(i) {
			t.Fatalf("slot reused before release: %+v", batches)
		}
		if i == 0 {
			first = batches[0]
		}
	}
	for _, invalid := range [][]MemoryBatch{
		{{Slot: MemorySlotCount, Generation: first.Generation, Length: first.Length}},
		{{Slot: first.Slot, Generation: first.Generation + 1, Length: first.Length}},
		{{Slot: first.Slot, Generation: first.Generation, Length: first.Length + 1}},
		{first, {Slot: MemorySlotCount, Generation: 1, Length: 1}},
		{first, first},
	} {
		pool.release(invalid)
		if batches, _ := pool.store([]Attachment{Text("overwritten")}); len(batches) != 0 {
			t.Fatal("invalid or partially valid release freed an occupied slot")
		}
	}
	pool.release([]MemoryBatch{first})
	next, _ := pool.store([]Attachment{Text("next")})
	if len(next) != 1 || next[0].Slot != first.Slot || next[0].Generation <= first.Generation {
		t.Fatalf("generation did not advance: %+v", next)
	}
	pool.release([]MemoryBatch{first})
	if batches, _ := pool.store([]Attachment{Text("overwritten")}); len(batches) != 0 {
		t.Fatal("stale release freed live bytes")
	}
	pool.release(next)
	pool.slots[first.Slot].generation = MemoryMaxGeneration
	if batches, _ := pool.store([]Attachment{Text("would wrap")}); len(batches) != 0 {
		t.Fatal("exhausted publication generation was reused")
	}
}

func TestPoolConcurrentReuse(t *testing.T) {
	pool := &memoryPool{mapping: memoryMapping{data: make([]byte, MemoryCapacity)}}
	var group sync.WaitGroup
	for worker := range 32 {
		group.Go(func() {
			for round := range 100 {
				text := fmt.Sprintf("snapshot-%d-%d-😀", worker, round)
				batches, _ := pool.store([]Attachment{Text(text)})
				if len(batches) == 0 {
					continue
				}
				runtime.Gosched()
				batch := batches[0]
				start := MemoryHeaderSize + int(batch.Slot)*MemorySlotSize
				if got := string(pool.mapping.data[start : start+int(batch.Length)]); got != text {
					t.Errorf("bytes overwritten before release: %q", got)
				}
				pool.release(batches)
			}
		})
	}
	group.Wait()
}

func TestPoolConcurrentClose(t *testing.T) {
	pool := &memoryPool{mapping: memoryMapping{data: make([]byte, MemoryCapacity)}}
	closed := 0
	pool.mapping.unmap = func() error {
		// Model invalidating mapped bytes. The race detector must see no
		// simultaneous publication/copy, and subsequent stores must fail.
		clear(pool.mapping.data[:MemoryHeaderSize+16])
		closed++
		return nil
	}
	var group sync.WaitGroup
	for range 8 {
		group.Go(func() {
			for range 100 {
				if batches, _ := pool.store([]Attachment{Text("snapshot")}); len(batches) != 0 {
					pool.release(batches)
				}
			}
		})
	}
	runtime.Gosched()
	if err := pool.close(); err != nil {
		t.Fatal(err)
	}
	group.Wait()
	if err := pool.close(); err != nil {
		t.Fatal(err)
	}
	if batches, _ := pool.store([]Attachment{Text("after close")}); len(batches) != 0 || closed != 1 {
		t.Fatal("closed mapping was used or unmapped twice")
	}
}

func TestPoolLaterCommitFailureDoesNotPublishOrConsumeSlots(t *testing.T) {
	fail := true
	commits := 0
	pool := &memoryPool{mapping: memoryMapping{
		data: make([]byte, MemoryHeaderSize+2*MemorySlotSize),
		commitSlot: func(slot int) error {
			commits++
			if slot == 1 && fail {
				return errors.New("commit unavailable")
			}
			return nil
		},
	}}
	values := []Attachment{Text(strings.Repeat("x", MemorySlotSize)), Text("y")}
	if batches, _ := pool.store(values); len(batches) != 0 {
		t.Fatal("failed later commit accepted a partial attachment set")
	}
	if pool.mapping.data[0] != 0 || pool.mapping.data[MemoryPublicationStride] != 0 || pool.mapping.data[MemoryHeaderSize] != 0 || pool.slots[0].busy || pool.slots[1].busy {
		t.Fatal("failed commit published, wrote bytes, or reserved a slot")
	}
	fail = false
	batches, _ := pool.store(values)
	if len(batches) != 2 || batches[0].Generation != 1 || batches[1].Generation != 1 {
		t.Fatalf("failed commit consumed a generation: %+v", batches)
	}
	pool.release(batches)
	before := commits
	if batches, _ := pool.store(values); len(batches) != 2 || commits != before {
		t.Fatal("slot backing was not retained across reuse")
	}
}
