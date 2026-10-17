package server

import (
	"context"

	api "github.com/web-infra-dev/rslint/internal/api"
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
		dispatch = NewPluginDispatcher(requester)
	}
	return h.handleLint(ctx, req, dispatch, requester)
}
