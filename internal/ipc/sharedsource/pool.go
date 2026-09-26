// Package sharedsource owns the Go writer side of the native source arena.
// It knows only byte storage and leases, not lint requests or IPC messages.
package sharedsource

import (
	"errors"
	"sync"
	"sync/atomic"
	"unsafe"
)

// Keep this layout in sync with rslint-native's source_transport module.
const slotCount = 16
const SlotSize = 16 * 1024 * 1024
const headerSize = 4096
const capacity = headerSize + slotCount*SlotSize

// Descriptor identifies an OS mapping owned by the native reader process.
type Descriptor struct {
	Version   uint32 `json:"version"`
	FD        int    `json:"fd,omitempty"`
	Handle    string `json:"handle,omitempty"`
	ProcessID uint32 `json:"processId,omitempty"`
}

// Batch identifies one published, immutable region until its release.
type Batch struct {
	Slot       uint32 `json:"slot"`
	Generation uint32 `json:"generation"`
	Length     uint32 `json:"length"`
}

type slot struct {
	generation uint32
	length     uint32
	busy       bool
}

// The platform mapping may reserve pages without committing their backing
// memory. Only Pool calls these operations, under its writer/close mutex.
type mappedSources struct {
	data       []byte
	commitSlot func(int) error
	unmap      func() error
}

// Pool owns the mapping and bounded slots. Its mutex also makes Close wait for
// any writer already copying bytes; no mapped slice escapes this package.
type Pool struct {
	mu      sync.Mutex
	mapping mappedSources
	slots   [slotCount]slot
}

func Open(descriptor Descriptor) (*Pool, error) {
	if descriptor.Version != 1 {
		return nil, errors.New("unsupported shared source version")
	}
	mapping, err := mapSources(descriptor)
	if err != nil {
		return nil, err
	}
	return &Pool{mapping: mapping}, nil
}

// Store copies all parts in order before atomically publishing the generation.
// False means unavailable capacity; the caller decides its transport fallback.
func (p *Pool) Store(parts []string) (Batch, bool) {
	if len(parts) == 0 {
		return Batch{}, false
	}
	size := 0
	for _, part := range parts {
		if len(part) > SlotSize-size {
			return Batch{}, false
		}
		size += len(part)
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mapping.data == nil {
		return Batch{}, false
	}
	for i := range p.slots {
		s := &p.slots[i]
		if s.busy || s.generation == ^uint32(0) {
			continue
		}
		if s.generation == 0 && p.mapping.commitSlot != nil {
			if err := p.mapping.commitSlot(i); err != nil {
				return Batch{}, false
			}
		}
		s.busy = true
		s.generation++
		s.length = uint32(size)
		start := headerSize + i*SlotSize
		offset := start
		for _, part := range parts {
			offset += copy(p.mapping.data[offset:start+size], part)
		}
		// Rust acquires this aligned word before admitting readers. A pipe
		// notification alone is not used as a cross-language memory fence.
		atomic.StoreUint32((*uint32)(unsafe.Pointer(&p.mapping.data[i*4])), s.generation)
		return Batch{Slot: uint32(i), Generation: s.generation, Length: s.length}, true
	}
	return Batch{}, false
}

// Release requires the native reader's explicit revocation acknowledgement.
// Missing acknowledgements leave slots occupied, bounding timeout retention.
func (p *Pool) Release(batch Batch) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if batch.Slot >= slotCount {
		return
	}
	s := &p.slots[batch.Slot]
	if s.generation == batch.Generation && s.length == batch.Length {
		s.busy = false
	}
}

func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.mapping.data == nil {
		return nil
	}
	var err error
	if p.mapping.unmap != nil {
		err = p.mapping.unmap()
	}
	p.mapping = mappedSources{}
	return err
}
