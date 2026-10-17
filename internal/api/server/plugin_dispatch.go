package server

import (
	"context"
	"fmt"
	api "github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// NewPluginDispatcher connects CLI/API plugin batches to their channel-owned
// output. LSP uses the linter's owned-frame adapter over its JSON-RPC route.
func NewPluginDispatcher(requester api.Requester) linter.EslintPluginDispatcher {
	return apiPluginDispatcher{requester: requester}
}

// The API transport owns output bytes; the linter sees only a batch allocator.
// Language projection does not import IPC or keep a mapping capability.
type apiPluginDispatcher struct{ requester api.Requester }

func (d apiPluginDispatcher) OpenBatch(ctx context.Context) (*linter.EslintPluginBatch, error) {
	output, err := d.requester.NewRequestOutput(ctx)
	if err != nil {
		return nil, err
	}
	batch := &linter.EslintPluginBatch{Release: func() { _ = output.Close() }}
	if output.Shared() {
		batch.Allocate = output.Allocate
		batch.Discard = output.Discard
	}
	batch.Dispatch = func(reqCtx context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
		req, frames := req.WithFrameAttachments()
		attachments := make([]ipc.Attachment, len(frames))
		for i, frame := range frames {
			attachments[i] = output.Bytes(frame)
		}
		msg, err := d.requester.SendRequest(reqCtx, api.KindPluginLint, req, attachments...)
		if err != nil {
			return nil, err
		}
		var result linter.EslintPluginLintResult
		if err := msg.Decode(&result); err != nil {
			return nil, fmt.Errorf("decode pluginLint result: %w", err)
		}
		return &result, nil
	}
	return batch, nil
}
