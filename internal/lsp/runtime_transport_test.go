package lsp

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/lsp/lsproto"
)

func TestRuntimeRequestUsesDedicatedTransport(t *testing.T) {
	called := false
	server := NewServer(&ServerOptions{
		Cwd: "/workspace",
		Err: io.Discard,
		RuntimeRequest: func(_ context.Context, method string, params any) (any, error) {
			called = true
			if method != "rslint/loadConfigs" {
				t.Fatalf("method = %q", method)
			}
			if params != "payload" {
				t.Fatalf("params = %#v", params)
			}
			return map[string]any{"ok": true}, nil
		},
	})

	result, err := server.sendRuntimeRequest(
		context.Background(),
		lsproto.Method("rslint/loadConfigs"),
		"payload",
	)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Fatal("private runtime requester was not called")
	}
	value, ok := result.(map[string]any)
	if !ok || value["ok"] != true {
		t.Fatalf("result = %#v", result)
	}
}

func TestRuntimeRequestRequiresDedicatedTransport(t *testing.T) {
	server := NewServer(&ServerOptions{Cwd: "/workspace", Err: io.Discard})
	_, err := server.sendRuntimeRequest(
		context.Background(),
		lsproto.Method("rslint/loadConfigs"),
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "private editor-runtime transport") {
		t.Fatalf("error = %v, want missing private transport", err)
	}
}

func TestRuntimeConfigFileWatchersCoverTopology(t *testing.T) {
	watchers := runtimeConfigFileWatchers("/workspace", true)
	patterns := make(map[string]bool, len(watchers))
	for _, watcher := range watchers {
		if watcher.GlobPattern.RelativePattern == nil {
			t.Fatalf("watcher does not use relative pattern: %#v", watcher)
		}
		patterns[watcher.GlobPattern.RelativePattern.Pattern] = true
	}
	for _, pattern := range []string{
		"**/rslint.config.js",
		"**/rslint.config.ts",
		"**/rslint.jsonc",
		"**/package.json",
		"**/npm-shrinkwrap.json",
		"**/pnpm-lock.yaml",
		"**/pnpm-workspace.yaml",
		"**/bun.lock",
		"**/.pnp.cjs",
		"**/.pnp.loader.mjs",
		"**/.pnp.data.json",
	} {
		if !patterns[pattern] {
			t.Errorf("missing watcher %q", pattern)
		}
	}
}
