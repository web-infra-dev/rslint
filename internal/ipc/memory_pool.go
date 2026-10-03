// Attachment storage is private to Channel. It owns bytes and publication
// leases, never application payloads or application request scheduling.
package ipc

import (
	"errors"
	"sync"
	"sync/atomic"
	"unsafe"
)

type memorySlot struct {
	generation uint32
	length     uint32
	busy       bool
}

type memoryMapping struct {
	data       []byte
	commitSlot func(int) error
	unmap      func() error
}

// memoryPool serializes reservation, copy, publication and close. Mapped
// slices never escape the package. One request may occupy multiple slots.
type memoryPool struct {
	mu      sync.Mutex
	mapping memoryMapping
	slots   [MemorySlotCount]memorySlot
}

func openMemoryPool(descriptor MemoryMapping) (*memoryPool, error) {
	if descriptor.Version != MemoryVersion {
		return nil, errors.New("unsupported shared memory version")
	}
	mapping, err := mapMemory(descriptor)
	if err != nil {
		return nil, err
	}
	return &memoryPool{mapping: mapping}, nil
}

// store plans complete attachments against the currently free slots without
// waiting or reserving partial work across requests. An attachment that cannot
// fit stays entirely inline; later smaller attachments may still fit. A value
// no larger than one slot never straddles slots. Larger values may span slots.
// All backing is committed and all bytes copied before any slot is published.
func (p *memoryPool) store(values []Attachment) ([]MemoryBatch, []*MemoryRange) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mapping.data == nil {
		return nil, nil
	}
	var free []int
	for i, slot := range p.slots {
		if !slot.busy && slot.generation != MemoryMaxGeneration {
			free = append(free, i)
		}
	}
	type planSlot struct{ slot, length int }
	var plan []planSlot
	ranges := make([]*MemoryRange, len(values))
	var total uint32
	for i, value := range values {
		size := value.length()
		if size == 0 {
			continue
		}
		tail := 0
		if len(plan) != 0 {
			tail = MemorySlotSize - plan[len(plan)-1].length
		}
		if size <= MemorySlotSize && size > tail {
			tail = 0 // Preserve a contiguous borrow for ordinary attachments.
		}
		available := tail + (len(free)-len(plan))*MemorySlotSize
		if size > available {
			continue
		}
		ranges[i] = &MemoryRange{Offset: total, Length: uint32(size)}
		total += uint32(size)
		remaining := size
		if tail > 0 {
			n := min(tail, remaining)
			plan[len(plan)-1].length += n
			remaining -= n
		}
		for remaining > 0 {
			n := min(MemorySlotSize, remaining)
			plan = append(plan, planSlot{slot: free[len(plan)], length: n})
			remaining -= n
		}
	}
	if len(plan) == 0 {
		return nil, nil
	}
	for _, item := range plan {
		if p.slots[item.slot].generation == 0 && p.mapping.commitSlot != nil {
			if err := p.mapping.commitSlot(item.slot); err != nil {
				return nil, nil
			}
		}
	}
	batchIndex, batchOffset := 0, 0
	for i, value := range values {
		if ranges[i] == nil {
			continue
		}
		for offset := 0; offset < value.length(); {
			item := plan[batchIndex]
			n := min(item.length-batchOffset, value.length()-offset)
			start := MemoryHeaderSize + item.slot*MemorySlotSize + batchOffset
			value.copyTo(p.mapping.data[start:start+n], offset)
			offset += n
			batchOffset += n
			if batchOffset == item.length {
				batchIndex++
				batchOffset = 0
			}
		}
	}
	batches := make([]MemoryBatch, len(plan))
	for i, item := range plan {
		slot := &p.slots[item.slot]
		slot.busy = true
		slot.generation++
		slot.length = uint32(item.length)
		atomic.StoreUint32((*uint32)(unsafe.Pointer(&p.mapping.data[item.slot*MemoryPublicationStride])), slot.generation)
		batches[i] = MemoryBatch{Slot: uint32(item.slot), Generation: slot.generation, Length: slot.length}
	}
	return batches, ranges
}

// release is all-or-nothing. Channel additionally requires the exact ordered
// list published for the matching request, including after caller cancellation.
func (p *memoryPool) release(batches []MemoryBatch) {
	p.mu.Lock()
	defer p.mu.Unlock()
	var seen uint32
	for _, batch := range batches {
		if batch.Slot >= MemorySlotCount || seen&(1<<batch.Slot) != 0 {
			return
		}
		seen |= 1 << batch.Slot
		slot := &p.slots[batch.Slot]
		if !slot.busy || slot.generation != batch.Generation || slot.length != batch.Length {
			return
		}
	}
	for _, batch := range batches {
		p.slots[batch.Slot].busy = false
	}
}

func (p *memoryPool) close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mapping.data == nil {
		return nil
	}
	var err error
	if p.mapping.unmap != nil {
		err = p.mapping.unmap()
	}
	p.mapping = memoryMapping{}
	return err
}
