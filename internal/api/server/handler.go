package server

import (
	"context"
	"fmt"

	api "github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// Handler implements rslint's concrete API requests.
type Handler struct{}

// HandleLint handles lint requests in IPC mode
func (h *Handler) HandleLint(req api.LintRequest) (*api.LintResponse, error) {
	return h.handleLint(context.Background(), req, nil, nil)
}

// HandleLintWithContext enables reverse pluginLint requests when Handler is
// hosted by the bidirectional API service. HandleLint remains available for
// direct callers that do not need community plugin execution.
func (h *Handler) HandleLintWithContext(ctx context.Context, req api.LintRequest, requester api.Requester) (*api.LintResponse, error) {
	var dispatch linter.EslintPluginDispatcher
	if requester != nil {
		dispatch = func(reqCtx context.Context, pluginReq linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
			var attachments []ipc.Attachment
			if peer, ok := requester.(api.PeerCapabilityRequester); ok && peer.PeerSupportsCapability(api.CapabilityPluginLintAttachments) {
				var texts []string
				pluginReq, texts = pluginReq.WithTextAttachments()
				attachments = make([]ipc.Attachment, len(texts))
				for i, text := range texts {
					attachments[i] = ipc.Text(text)
				}
			}
			msg, err := requester.SendRequest(reqCtx, api.KindPluginLint, pluginReq, attachments...)
			if err != nil {
				return nil, err
			}
			var result linter.EslintPluginLintResult
			if err := msg.Decode(&result); err != nil {
				return nil, fmt.Errorf("decode pluginLint result: %w", err)
			}
			return &result, nil
		}
	}
	return h.handleLint(ctx, req, dispatch, requester)
}
