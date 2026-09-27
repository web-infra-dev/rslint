// Package pluginlint adapts immutable lint requests to the CLI IPC peer.
// Transport storage, platform resources and attachment leases belong to ipc.
package pluginlint

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
)

type sourceFile struct {
	linter.EslintPluginLintFile
	SourceIndex *uint32 `json:"sourceIndex,omitempty"`
}

type sourceRequest struct {
	linter.EslintPluginLintRequest
	Files []sourceFile `json:"files"`
}

// This consumer-owned seam keeps business batching tests independent of OS
// mappings. The production implementation is the existing IPC channel.
type requestChannel interface {
	AttachmentLimit() int
	SendRequest(ctx context.Context, kind ipc.MessageKind, payload any, attachments ...string) (*ipc.Message, error)
}

// Client preserves logical request metadata and ordering while bounding the
// plugin requests in flight across all logical batches.
type Client struct {
	channel  requestChannel
	inFlight chan struct{}
}

// New binds the plugin application protocol to an existing IPC channel.
func New(channel *ipc.Channel) *Client {
	return &Client{channel: channel, inFlight: make(chan struct{}, 8)}
}

// Dispatch preserves logical batch semantics across the transport's byte budget.
func (d *Client) Dispatch(ctx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
	limit := d.channel.AttachmentLimit()
	if limit == 0 || len(req.Files) == 0 {
		return d.send(ctx, req)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	// The byte budget must not become an execution barrier: a slow file in one
	// part must not prevent later parts from reaching idle plugin workers.
	// Individual oversized files stay on the inline path.
	var parts []linter.EslintPluginLintRequest
	for start := 0; start < len(req.Files); {
		end, size := start, 0
		for end < len(req.Files) {
			cost := 0
			if text := req.Files[end].Text; text != nil && len(*text) <= limit {
				cost = len(*text)
			}
			if cost > limit-size {
				break
			}
			size += cost
			end++
		}
		part := req
		part.Files = req.Files[start:end]
		parts = append(parts, part)
		start = end
	}
	responses := make([]struct {
		result *linter.EslintPluginLintResult
		err    error
	}, len(parts))
	var group sync.WaitGroup
	started := 0
dispatchParts:
	for i, part := range parts {
		if ctx.Err() != nil {
			break
		}
		// Acquire before spawning: waiting parts retain no goroutine, and all
		// logical requests share this limit rather than multiplying concurrency.
		select {
		case d.inFlight <- struct{}{}:
		case <-ctx.Done():
			break dispatchParts
		}
		started++
		group.Go(func() {
			defer func() { <-d.inFlight }()
			defer func() {
				// Preserve the linter's dispatch panic isolation inside this new
				// goroutine boundary; a bad request must not crash the process.
				if recovered := recover(); recovered != nil {
					responses[i].err = fmt.Errorf("eslint-plugin dispatch panicked: %v", recovered)
				}
			}()
			responses[i].result, responses[i].err = d.send(ctx, part)
		})
	}
	// A cancelled caller still owns its snapshots until every started writer
	// and request has returned. Join before returning on either errors or cancel.
	group.Wait()
	var result linter.EslintPluginLintResult
	var cancelled error
	for _, response := range responses[:started] {
		if response.err != nil {
			// Match the linter's failure policy: cancellation must not hide a
			// real failure in a later part. Within each class keep input order.
			if !errors.Is(response.err, context.Canceled) {
				return nil, response.err
			}
			if cancelled == nil {
				cancelled = response.err
			}
			continue
		}
		result.Results = append(result.Results, response.result.Results...)
	}
	if cancelled != nil {
		return nil, cancelled
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return &result, nil
}

func (d *Client) send(ctx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	wire, attachments := encode(req)
	msg, err := d.channel.SendRequest(ctx, "pluginLint", wire, attachments...)
	if err != nil {
		return nil, err
	}
	var result linter.EslintPluginLintResult
	if err := msg.Decode(&result); err != nil {
		return nil, fmt.Errorf("decode pluginLint result: %w", err)
	}
	return &result, nil
}

// Attachments are complete immutable snapshots. IPC alone decides how to
// carry them; this adapter only assigns a stable index to each source file.
func encode(req linter.EslintPluginLintRequest) (sourceRequest, []string) {
	files := make([]sourceFile, len(req.Files))
	attachments := make([]string, 0, len(req.Files))
	for i, file := range req.Files {
		files[i].EslintPluginLintFile = file
		if file.Text == nil {
			continue
		}
		index := uint32(len(attachments))
		attachments = append(attachments, *file.Text)
		files[i].Text = nil
		files[i].SourceIndex = &index
	}
	return sourceRequest{EslintPluginLintRequest: req, Files: files}, attachments
}
