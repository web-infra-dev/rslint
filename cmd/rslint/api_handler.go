package main

import (
	"context"
	"fmt"
	"os"

	api "github.com/web-infra-dev/rslint/internal/api"
	"github.com/web-infra-dev/rslint/internal/linter"
)

// IPCHandler implements the ipc.Handler interface
type IPCHandler struct{}

// HandleLint handles lint requests in IPC mode
func (h *IPCHandler) HandleLint(req api.LintRequest) (*api.LintResponse, error) {
	return h.handleLint(context.Background(), req, nil, nil)
}

// HandleLintWithContext enables reverse pluginLint requests when IPCHandler is
// hosted by the bidirectional API service. HandleLint remains available for
// direct callers that do not need community plugin execution.
func (h *IPCHandler) HandleLintWithContext(ctx context.Context, req api.LintRequest, requester api.Requester) (*api.LintResponse, error) {
	var dispatch linter.EslintPluginDispatcher
	if requester != nil {
		dispatch = func(reqCtx context.Context, pluginReq linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
			msg, err := requester.SendRequest(reqCtx, api.KindPluginLint, pluginReq)
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

// runAPI runs the linter in IPC mode
func runAPI() int {
	handler := &IPCHandler{}
	service := api.NewService(os.Stdin, os.Stdout, handler)

	if err := service.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "error in IPC mode: %v\n", err)
		return 1
	}
	return 0
}
