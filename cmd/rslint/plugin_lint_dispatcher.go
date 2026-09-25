//go:build !js

package main

import (
	"context"
	"fmt"
	"unicode/utf8"

	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/ipc/sharedsource"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// This adapter alone translates linter snapshots into the private CLI wire
// format. The shared source pool never sees a lint request or an IPC channel.
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
	Files       []pluginSourceFile  `json:"files"`
	SourceBatch *sharedsource.Batch `json:"sourceBatch,omitempty"`
}

type pluginSourceResponse struct {
	linter.EslintPluginLintResult
	ReleasedSource *sharedsource.Batch `json:"releasedSource,omitempty"`
}

// The consumer owns this small boundary so wire tests need no OS mapping.
type pluginSourceStore interface {
	Store(parts []string) (sharedsource.Batch, bool)
	Release(batch sharedsource.Batch)
	Close() error
}

type pluginLintDispatcher struct {
	channel *ipc.Channel
	sources pluginSourceStore
}

func newPluginLintDispatcher(channel *ipc.Channel, descriptor *sharedsource.Descriptor) *pluginLintDispatcher {
	d := &pluginLintDispatcher{channel: channel}
	if descriptor != nil {
		// Sharing is optional. Mapping failures keep complete inline text;
		// this transport policy belongs here, not in the memory pool.
		if pool, err := sharedsource.Open(*descriptor); err == nil {
			d.sources = pool
		}
	}
	return d
}

func (d *pluginLintDispatcher) close() {
	if d.sources != nil {
		_ = d.sources.Close()
	}
}

func (d *pluginLintDispatcher) dispatch(ctx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
	if d.sources == nil || len(req.Files) == 0 {
		return d.send(ctx, req)
	}
	// Split logical batches by source bytes, retaining the original result order
	// and rules/configuration. Individual oversized files stay on the inline path.
	var result linter.EslintPluginLintResult
	for start := 0; start < len(req.Files); {
		end, size := start, 0
		for end < len(req.Files) {
			cost := 0
			if text := req.Files[end].Text; text != nil && len(*text) <= sharedsource.SlotSize {
				cost = len(*text)
			}
			if cost > sharedsource.SlotSize-size {
				break
			}
			size += cost
			end++
		}
		part := req
		part.Files = req.Files[start:end]
		response, err := d.send(ctx, part)
		if err != nil {
			return nil, err
		}
		result.Results = append(result.Results, response.Results...)
		start = end
	}
	return &result, nil
}

func (d *pluginLintDispatcher) send(ctx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	wire, batch := d.encode(req)
	msg, err := d.channel.SendRequest(ctx, kindPluginLint, wire)
	if err != nil {
		return nil, err
	}
	var response pluginSourceResponse
	if err := msg.Decode(&response); err != nil {
		return nil, fmt.Errorf("decode pluginLint result: %w", err)
	}
	// Match the acknowledgement to THIS request before allowing reuse. A normal
	// result alone says nothing about an earlier native reader still in flight.
	if batch != nil && response.ReleasedSource != nil && *response.ReleasedSource == *batch {
		d.sources.Release(*batch)
	}
	return &response.EslintPluginLintResult, nil
}

func (d *pluginLintDispatcher) encode(req linter.EslintPluginLintRequest) (any, *sharedsource.Batch) {
	if d.sources == nil {
		return req, nil
	}
	files := make([]pluginSourceFile, len(req.Files))
	parts := make([]string, 0, len(req.Files))
	var length uint32
	for i, file := range req.Files {
		files[i].EslintPluginLintFile = file
		if file.Text == nil || len(*file.Text) > sharedsource.SlotSize-int(length) || !utf8.ValidString(*file.Text) {
			continue
		}
		size := uint32(len(*file.Text))
		parts = append(parts, *file.Text)
		files[i].Text = nil
		files[i].SourceRange = &pluginSourceRange{Offset: length, Length: size}
		length += size
	}
	if batch, ok := d.sources.Store(parts); ok {
		return pluginSourceRequest{EslintPluginLintRequest: req, Files: files, SourceBatch: &batch}, &batch
	}
	// The original request is unchanged, including every complete source string.
	return req, nil
}
