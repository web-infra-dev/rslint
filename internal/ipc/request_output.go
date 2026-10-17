package ipc

import (
	"context"
	"errors"
	"sync/atomic"
	"unsafe"
)

type outputBlock struct {
	first, slots, used int
}

type outputRange struct {
	block, offset, length int
}

// RequestOutput owns unpublished byte storage for one outbound request. A single
// producer writes it synchronously; Close must follow SendRequest on every path.
// Allocate/Bytes views are consumed by publication and must not be read or
// written afterwards, including while SendRequest is waiting for the response.
// Shared bytes are written directly into channel storage. Publication transfers
// their slot lifetime to the channel's existing reader acknowledgement protocol.
// This type contains no application payload or encoding policy.
type RequestOutput struct {
	ctx               context.Context
	pool              *memoryPool
	blocks            []outputBlock
	ranges            map[*byte]outputRange
	order             []*byte
	published, closed bool
	writer            bool
}

// NewRequestOutput supplies owned bytes when no channel storage is available.
// Embedded requesters use this same explicit output contract.
func NewRequestOutput(ctx context.Context) *RequestOutput {
	return &RequestOutput{ctx: ctx}
}

func (c *Channel) NewRequestOutput(ctx context.Context) (*RequestOutput, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil, c.closeErr
	}
	out := NewRequestOutput(ctx)
	if p := c.memory; p != nil {
		p.mu.Lock()
		defer p.mu.Unlock()
		if !p.closing && p.mapping.data != nil {
			p.writers++
			out.pool = p
			out.writer = true
			out.ranges = make(map[*byte]outputRange)
		}
	}
	return out, nil
}

// Shared reports whether this output can allocate in the negotiated mapping.
// An unavailable mapping keeps the caller's existing owned-buffer allocator.
func (o *RequestOutput) Shared() bool { return o.pool != nil }

// Allocate returns exactly size writable bytes. No view may be used after Close.
// A full/fragmented mapping falls back to owned bytes for this allocation.
func (o *RequestOutput) Allocate(size int) []byte {
	if err := o.ctx.Err(); err != nil {
		panic(err)
	}
	if o.closed || o.published {
		panic(errors.New("ipc: output is no longer writable"))
	}
	if size < 0 || size > MaxFrameSize {
		panic(errors.New("ipc: invalid output size"))
	}
	p := o.pool
	if p == nil || size == 0 {
		return make([]byte, size)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.closing {
		panic(errors.New("ipc: channel output is closed"))
	}
	block := len(o.blocks) - 1
	if block >= 0 {
		b := &o.blocks[block]
		tail := b.slots*MemorySlotSize - b.used
		if size > tail {
			extra := (size - tail + MemorySlotSize - 1) / MemorySlotSize
			if p.reserveOutput(b.first+b.slots, extra) {
				b.slots += extra
			}
		}
	}
	if block < 0 || size > o.blocks[block].slots*MemorySlotSize-o.blocks[block].used {
		count := (size + MemorySlotSize - 1) / MemorySlotSize
		first, run := -1, 0
		for i, slot := range p.slots {
			if MemoryHeaderSize+(i+1)*MemorySlotSize > len(p.mapping.data) || slot.busy || slot.reserved || slot.generation == MemoryMaxGeneration {
				run = 0
				continue
			}
			run++
			if run == count {
				first = i - count + 1
				break
			}
		}
		if first < 0 {
			return make([]byte, size)
		}
		if !p.reserveOutput(first, count) {
			return make([]byte, size)
		}
		o.blocks = append(o.blocks, outputBlock{first: first, slots: count})
		block++
	}
	b := &o.blocks[block]
	start := MemoryHeaderSize + b.first*MemorySlotSize + b.used
	data := p.mapping.data[start : start+size : start+size]
	o.ranges[unsafe.SliceData(data)] = outputRange{block: block, offset: b.used, length: size}
	o.order = append(o.order, unsafe.SliceData(data))
	b.used += size
	return data
}

// reserveOutput is called under the pool lock. Commit all pages before
// reserving any slot, so allocation failure leaves no partial reservation.
func (p *memoryPool) reserveOutput(first, count int) bool {
	if first < 0 || count <= 0 || first+count > len(p.slots) || MemoryHeaderSize+(first+count)*MemorySlotSize > len(p.mapping.data) {
		return false
	}
	for i := first; i < first+count; i++ {
		slot := p.slots[i]
		if slot.busy || slot.reserved || slot.generation == MemoryMaxGeneration {
			return false
		}
	}
	for i := first; i < first+count; i++ {
		if p.slots[i].generation == 0 && p.mapping.commitSlot != nil {
			if err := p.mapping.commitSlot(i); err != nil {
				return false
			}
		}
	}
	for i := first; i < first+count; i++ {
		p.slots[i].reserved = true
	}
	return true
}

// Discard rolls back the most recent failed encoding. Successful allocations
// must appear exactly once, in allocation order, in the outbound attachments.
func (o *RequestOutput) Discard(data []byte) {
	if o.closed || o.published {
		panic(errors.New("ipc: output is no longer writable"))
	}
	key := unsafe.SliceData(data)
	span, ok := o.ranges[key]
	if !ok {
		return
	} // An owned fallback has no reserved slot.
	if len(o.order) == 0 || o.order[len(o.order)-1] != key {
		panic(errors.New("ipc: only the last output can be discarded"))
	}
	p := o.pool
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(o.ranges, key)
	o.order[len(o.order)-1] = nil
	o.order = o.order[:len(o.order)-1]
	b := &o.blocks[span.block]
	b.used = span.offset
	slots := (b.used + MemorySlotSize - 1) / MemorySlotSize
	for i := b.first + slots; i < b.first+b.slots; i++ {
		p.slots[i].reserved = false
	}
	b.slots = slots
	if b.used == 0 {
		o.blocks = o.blocks[:span.block]
	}
}

// Bytes binds exactly an allocation returned by this output to its publication.
// Independently owned/frozen bytes keep the normal byte attachment path.
func (o *RequestOutput) Bytes(data []byte) Attachment {
	if o.closed || o.published {
		panic(errors.New("ipc: output is no longer writable"))
	}
	a := Bytes(data)
	if span, ok := o.ranges[unsafe.SliceData(data)]; ok && span.length == len(data) {
		a.output = o
	}
	return a
}

func (o *RequestOutput) publish(msg *Message, p *memoryPool, values []Attachment) ([]MemoryBatch, error) {
	if o.closed || o.published || p == nil || p != o.pool {
		return nil, errors.New("ipc: invalid output owner")
	}
	p.mu.Lock()
	if p.closing {
		p.mu.Unlock()
		return nil, errors.New("ipc: channel output is closed")
	}
	msg.Attachments = make([]AttachmentData, len(values))
	bases := make([]uint32, len(o.blocks))
	var batches []MemoryBatch
	var total uint32
	for i, b := range o.blocks {
		bases[i] = total
		for n, remaining := 0, b.used; remaining > 0; n++ {
			length := min(remaining, MemorySlotSize)
			slot := &p.slots[b.first+n]
			if !slot.reserved || slot.busy {
				p.mu.Unlock()
				return nil, errors.New("ipc: output reservation lost")
			}
			batches = append(batches, MemoryBatch{Slot: uint32(b.first + n), Generation: slot.generation + 1, Length: uint32(length)})
			remaining -= length
		}
		total += uint32(b.used)
	}
	var end uint32
	for i, a := range values {
		if a.output == nil {
			msg.Attachments[i] = a.inline()
			continue
		}
		span, ok := o.ranges[unsafe.SliceData(a.bytes)]
		if !ok || a.output != o || len(a.bytes) != span.length {
			p.mu.Unlock()
			return nil, errors.New("ipc: invalid output attachment")
		}
		offset := bases[span.block] + uint32(span.offset)
		if offset != end {
			p.mu.Unlock()
			return nil, errors.New("ipc: output attachments are out of allocation order")
		}
		end += uint32(span.length)
		msg.Attachments[i] = AttachmentData{Range: &MemoryRange{Offset: offset, Length: uint32(span.length)}}
	}
	if end != total {
		p.mu.Unlock()
		return nil, errors.New("ipc: output has unused allocations")
	}
	for _, b := range batches {
		slot := &p.slots[b.Slot]
		slot.reserved = false
		slot.busy = true
		slot.generation = b.Generation
		slot.length = b.Length
		atomic.StoreUint32((*uint32)(unsafe.Pointer(&p.mapping.data[int(b.Slot)*MemoryPublicationStride])), b.Generation)
	}
	o.published = true
	// Publication ends producer access. Only reader acknowledgement owns the
	// occupied slots now; a channel shutdown need not wait for the Node rules.
	o.writer = false
	p.writers--
	clear(o.ranges)
	clear(o.order)
	o.order = nil
	p.mu.Unlock()
	msg.Transport = &TransportMetadata{Batches: batches}
	return batches, nil
}

// Close aborts unpublished reservations or ends the producer's mapping borrow.
// Published slots remain owned by Channel until an exact reader acknowledgement.
func (o *RequestOutput) Close() error {
	if o.closed {
		return nil
	}
	o.closed = true
	p := o.pool
	clear(o.ranges)
	clear(o.order)
	o.order = nil
	if p == nil {
		return nil
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !o.published {
		for _, b := range o.blocks {
			for i := b.first; i < b.first+b.slots; i++ {
				p.slots[i].reserved = false
			}
		}
	}
	if o.writer {
		p.writers--
		o.writer = false
	}
	if p.closing {
		return p.closeMapping()
	}
	return nil
}
