//go:build !js

package main

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"unicode/utf8"
	"unsafe"

	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// This is a CLI transport detail, not part of the linter or API/LSP protocol.
// Keep the layout in sync with rslint-native's source_transport module.
const pluginSourceSlots = 16
const pluginSourceSlotSize = 16 * 1024 * 1024
const pluginSourceHeaderSize = 4096
const pluginSourceCapacity = pluginSourceHeaderSize + pluginSourceSlots*pluginSourceSlotSize

type pluginSourceMapping struct {
	Version   uint32 `json:"version"`
	FD        int    `json:"fd,omitempty"`
	Handle    string `json:"handle,omitempty"`
	ProcessID uint32 `json:"processId,omitempty"`
}

type pluginSourceBatch struct {
	Slot       uint32 `json:"slot"`
	Generation uint32 `json:"generation"`
	Length     uint32 `json:"length"`
}

type pluginSourceRange struct {
	Offset uint32 `json:"offset"`
	Length uint32 `json:"length"`
}

type pluginSourceFile struct {
	linter.EslintPluginLintFile
	SourceRange *pluginSourceRange `json:"sourceRange,omitempty"`
}

type pluginSourceRequest struct {
	linter.EslintPluginLintRequest
	Files       []pluginSourceFile `json:"files"`
	SourceBatch *pluginSourceBatch `json:"sourceBatch,omitempty"`
}

type pluginSourceResponse struct {
	linter.EslintPluginLintResult
	ReleasedSource *pluginSourceBatch `json:"releasedSource,omitempty"`
}

type pluginSourceSlot struct {
	generation uint32
	busy       bool
}

type pluginSourcePool struct {
	data  []byte
	close func() error
	mu    sync.Mutex
	slots [pluginSourceSlots]pluginSourceSlot
}

func openPluginSourcePool(descriptor *pluginSourceMapping) *pluginSourcePool {
	if descriptor == nil || descriptor.Version != 1 {
		return nil
	}
	data, closeMapping, err := mapPluginSources(*descriptor)
	if err != nil {
		// Shared memory is an optimization. Preserve the complete original
		// source through the existing JSON transport if mapping is unavailable.
		return nil
	}
	return &pluginSourcePool{data: data, close: closeMapping}
}

// encode copies immutable snapshots before publishing their ranges. Fixed slots
// bound retained memory even if a peer disconnects or a native reader times out.
// Oversized/invalid-UTF-8 strings stay inline, retaining encoding/json semantics.
func (p *pluginSourcePool) encode(req linter.EslintPluginLintRequest) (any, *pluginSourceBatch) {
	if p == nil {
		return req, nil
	}
	p.mu.Lock()
	index := -1
	for i := range p.slots {
		s := &p.slots[i]
		if !s.busy && s.generation != ^uint32(0) {
			s.busy = true
			s.generation++
			index = i
			break
		}
	}
	if index < 0 {
		p.mu.Unlock()
		return req, nil
	}
	batch := pluginSourceBatch{Slot: uint32(index), Generation: p.slots[index].generation}
	p.mu.Unlock()

	files := make([]pluginSourceFile, len(req.Files))
	start := pluginSourceHeaderSize + index*pluginSourceSlotSize
	buffer := p.data[start : start+pluginSourceSlotSize]
	shared := false
	for i, file := range req.Files {
		files[i].EslintPluginLintFile = file
		if file.Text == nil || len(*file.Text) > len(buffer)-int(batch.Length) || !utf8.ValidString(*file.Text) {
			continue
		}
		size := copy(buffer[batch.Length:], *file.Text)
		files[i].Text = nil
		files[i].SourceRange = &pluginSourceRange{Offset: batch.Length, Length: uint32(size)}
		batch.Length += uint32(size)
		shared = true
	}
	if !shared {
		p.release(batch)
		return req, nil
	}
	// Publish with a lock-free, aligned 32-bit atomic. Rust acquires this word
	// before admitting readers; correctness does not depend on a pipe syscall
	// serving as a memory fence on every supported CPU.
	atomic.StoreUint32((*uint32)(unsafe.Pointer(&p.data[index*4])), batch.Generation)
	return pluginSourceRequest{EslintPluginLintRequest: req, Files: files, SourceBatch: &batch}, &batch
}

func (p *pluginSourcePool) release(batch pluginSourceBatch) {
	p.mu.Lock()
	defer p.mu.Unlock()
	s := &p.slots[batch.Slot]
	if s.generation == batch.Generation {
		s.busy = false
	}
}

func (p *pluginSourcePool) dispatch(ch *ipc.Channel) linter.EslintPluginDispatcher {
	send := func(ctx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		wire, batch := p.encode(req)
		msg, err := ch.SendRequest(ctx, kindPluginLint, wire)
		if err != nil {
			return nil, err
		}
		var response pluginSourceResponse
		if err := msg.Decode(&response); err != nil {
			return nil, fmt.Errorf("decode pluginLint result: %w", err)
		}
		// A normal lint result is NOT a read-completion fence. Only Rust's
		// explicit revocation acknowledgement permits another writer here.
		if batch != nil && response.ReleasedSource != nil && *response.ReleasedSource == *batch {
			p.release(*batch)
		}
		return &response.EslintPluginLintResult, nil
	}
	return func(ctx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		if p == nil || len(req.Files) == 0 {
			return send(ctx, req)
		}
		// A logical lint batch can contain arbitrarily many snapshots. Split
		// its transport requests by bytes, retaining result order and the same
		// rules/configuration. A single oversized file keeps the inline path.
		var result linter.EslintPluginLintResult
		for start := 0; start < len(req.Files); {
			end, size := start, 0
			for end < len(req.Files) {
				cost := 0
				if text := req.Files[end].Text; text != nil && len(*text) <= pluginSourceSlotSize {
					cost = len(*text)
				}
				if cost > pluginSourceSlotSize-size {
					break
				}
				size += cost
				end++
			}
			part := req
			part.Files = req.Files[start:end]
			response, err := send(ctx, part)
			if err != nil {
				return nil, err
			}
			result.Results = append(result.Results, response.Results...)
			start = end
		}
		return &result, nil
	}
}
