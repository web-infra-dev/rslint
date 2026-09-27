// Source attachment storage is private to Channel. It owns only bytes and
// publication leases, never application payloads.
package ipc

import (
	"errors"
	"sync"
	"sync/atomic"
	"unsafe"
)

type sourceSlot struct {
	generation uint32
	length     uint32
	busy       bool
}

// The platform mapping may reserve pages without committing their backing
// memory. Only sourcePool calls these operations, under its writer/close mutex.
type sourceMapping struct {
	data       []byte
	commitSlot func(int) error
	unmap      func() error
}

// sourcePool owns the mapping and bounded slots. Its mutex also makes close wait for
// any writer already copying bytes; no mapped slice escapes this package.
type sourcePool struct {
	mu      sync.Mutex
	mapping sourceMapping
	slots   [SourceSlotCount]sourceSlot
}

func openSourcePool(descriptor SourceDescriptor) (*sourcePool, error) {
	if descriptor.Version != SourceVersion {
		return nil, errors.New("unsupported shared source version")
	}
	mapping, err := mapSourceMapping(descriptor)
	if err != nil {
		return nil, err
	}
	return &sourcePool{mapping: mapping}, nil
}

// store copies all parts in order before atomically publishing the generation.
// False means unavailable capacity; the caller decides its transport fallback.
func (p *sourcePool) store(parts []string) (SourceBatch, bool) {
	if len(parts) == 0 {
		return SourceBatch{}, false
	}
	size := 0
	for _, part := range parts {
		if len(part) > SourceSlotSize-size {
			return SourceBatch{}, false
		}
		size += len(part)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mapping.data == nil {
		return SourceBatch{}, false
	}
	for i := range p.slots {
		s := &p.slots[i]
		if s.busy || s.generation == SourceMaxGeneration {
			continue
		}
		if s.generation == 0 && p.mapping.commitSlot != nil {
			if err := p.mapping.commitSlot(i); err != nil {
				return SourceBatch{}, false
			}
		}
		s.busy = true
		s.generation++
		s.length = uint32(size)
		start := SourceHeaderSize + i*SourceSlotSize
		offset := start
		for _, part := range parts {
			offset += copy(p.mapping.data[offset:start+size], part)
		}
		// Rust acquires this aligned word before admitting readers. A pipe
		// notification alone is not used as a cross-language memory fence.
		atomic.StoreUint32((*uint32)(unsafe.Pointer(&p.mapping.data[i*SourcePublicationStride])), s.generation)
		return SourceBatch{Slot: uint32(i), Generation: s.generation, Length: s.length}, true
	}
	return SourceBatch{}, false
}

// release requires the native reader's explicit revocation acknowledgement.
// Missing acknowledgements leave slots occupied, bounding timeout retention.
func (p *sourcePool) release(batch SourceBatch) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if batch.Slot >= SourceSlotCount {
		return
	}
	s := &p.slots[batch.Slot]
	if s.generation == batch.Generation && s.length == batch.Length {
		s.busy = false
	}
}

func (p *sourcePool) close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mapping.data == nil {
		return nil
	}
	var err error
	if p.mapping.unmap != nil {
		err = p.mapping.unmap()
	}
	p.mapping = sourceMapping{}
	return err
}
