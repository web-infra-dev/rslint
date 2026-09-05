//go:build !js

package main

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/microsoft/typescript-go/shim/bundled"
	"github.com/microsoft/typescript-go/shim/tspath"
	"github.com/microsoft/typescript-go/shim/vfs/osvfs"
	"github.com/web-infra-dev/rslint/internal/config"
	"github.com/web-infra-dev/rslint/internal/config/discovery"
	"github.com/web-infra-dev/rslint/internal/ipc"
	"github.com/web-infra-dev/rslint/internal/linter"
)

func TestCLIConfigActivationBeforeExecution(t *testing.T) {
	for _, test := range []struct {
		name          string
		ignore        bool
		disabled      bool
		typeCheckOnly bool
		badOptions    bool
		badProject    bool
		fix           bool
		succeed       bool
		defaultFormat bool
	}{
		{name: "lint failure"},
		{name: "no targets", ignore: true},
		{name: "disabled plugin", disabled: true},
		{name: "type check only", typeCheckOnly: true},
		{name: "preflight failure", badOptions: true},
		{name: "Program failure", badProject: true},
		{name: "fix failure", fix: true},
		{name: "successful activation", succeed: true},
		{name: "default format abort", defaultFormat: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			targetPath := tspath.NormalizePath(filepath.Join(dir, "index.ts"))
			const content = "const value = true; if (!!value) { debugger; }\n"
			if err := os.WriteFile(targetPath, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "tsconfig.json"), []byte(`{"files":["index.ts"]}`), 0o644); err != nil {
				t.Fatal(err)
			}
			project := "./tsconfig.json"
			if test.badProject {
				project = "./missing.json"
			}
			entries := config.RslintConfig{{
				Files:   []string{"**/*.ts"},
				Plugins: []string{"external"},
				Rules: config.Rules{
					"no-debugger":           "error",
					"no-extra-boolean-cast": "error",
					"external/check":        "error",
				},
				LanguageOptions: &config.LanguageOptions{ParserOptions: &config.ParserOptions{Project: config.ProjectPaths{project}}},
			}}
			if test.disabled {
				entries[0].Rules["external/check"] = "off"
			}
			if test.badOptions {
				entries[0].Rules["no-debugger"] = []any{"error", "invalid-option"}
			}
			if test.ignore {
				entries = append(config.RslintConfig{{Ignores: []string{"**/*.ts"}}}, entries...)
			}
			catalog := explicitConfigCatalogForTest(dir, entries)
			catalog.EslintPlugins = []config.EslintPluginEntry{{Prefix: "external", RuleNames: []string{"check"}}}
			fsys := &commandReadCountingFS{FS: bundled.WrapFS(osvfs.FS()), reads: make(map[string]int)}
			calls, dispatches := 0, 0
			activationError := errors.New("worker preparation rejected")
			format := "jsonline"
			if test.defaultFormat {
				format = "default"
			}
			code, stdout, stderr := runLintCommandWithDispatcherForTest(t, dir, lintArgs{
				ConfigCatalog:  catalog,
				FS:             fsys,
				AllowFiles:     []string{targetPath},
				Format:         format,
				SingleThreaded: true,
				TypeCheck:      test.typeCheckOnly,
				TypeCheckOnly:  test.typeCheckOnly,
				Fix:            test.fix,
				CompleteConfigActivation: func() error {
					calls++
					if !test.ignore && !test.badOptions && !test.badProject && fsys.readCount(targetPath) == 0 {
						t.Error("activation joined before Program read the target")
					}
					if dispatches != 0 {
						t.Error("plugin execution preceded activation")
					}
					if test.succeed {
						return nil
					}
					return activationError
				},
			}, func(_ context.Context, req linter.EslintPluginLintRequest) (*linter.EslintPluginLintResult, error) {
				dispatches++
				if calls != 1 {
					t.Error("plugin dispatch did not await activation")
				}
				results := make([]linter.EslintPluginFileResult, len(req.Files))
				for index, file := range req.Files {
					results[index].FilePath = file.Path
				}
				return &linter.EslintPluginLintResult{Results: results}, nil
			})
			if calls != 1 {
				t.Fatalf("activation joins = %d, want 1", calls)
			}
			if test.succeed {
				if dispatches != 1 || !strings.Contains(stdout, "no-debugger") || stderr != "" {
					t.Fatalf("execution after activation: dispatches=%d stdout=%q stderr=%q", dispatches, stdout, stderr)
				}
			} else if test.defaultFormat {
				if code != 1 || dispatches != 0 || !strings.Contains(stdout, activationError.Error()) || strings.Contains(stdout, "no-debugger") || stderr != "" {
					t.Fatalf("default activation failure: code=%d dispatches=%d stdout=%q stderr=%q", code, dispatches, stdout, stderr)
				}
			} else if code != 1 || dispatches != 0 || stdout != "" || strings.Count(stderr, activationError.Error()) != 1 {
				t.Fatalf("activation failure: code=%d dispatches=%d stdout=%q stderr=%q", code, dispatches, stdout, stderr)
			}
			if got, err := os.ReadFile(targetPath); err != nil || string(got) != content {
				t.Fatalf("source changed before activation: content=%q err=%v", got, err)
			}
		})
	}
}

func TestCLIConfigPreparationProtocol(t *testing.T) {
	for _, test := range []struct {
		name     string
		warmup   bool
		plugins  bool
		mismatch bool
		fail     bool
	}{
		{name: "plugins warm up", warmup: true, plugins: true},
		{name: "native only", warmup: true},
		{name: "single threaded", plugins: true},
		{name: "changed metadata", warmup: true, plugins: true, mismatch: true},
		{name: "worker failure", warmup: true, plugins: true, fail: true},
	} {
		t.Run(test.name, func(t *testing.T) {
			cli, peer := newCLIChannelPair(t)
			var kinds []ipc.MessageKind
			peer.SetInboundHandler(func(_ context.Context, msg *ipc.Message) (any, error) {
				kinds = append(kinds, msg.Kind)
				var request discovery.ConfigActivationRequest
				if err := msg.Decode(&request); err != nil {
					return nil, err
				}
				response := discovery.ConfigActivationResponse{TransactionID: request.TransactionID}
				if test.plugins {
					response.EslintPluginEntries = []config.EslintPluginEntry{{Prefix: "external", RuleNames: []string{"check"}}}
				}
				if msg.Kind == kindActivateConfigs {
					if test.fail {
						return nil, errors.New("worker preparation rejected")
					}
					if test.mismatch {
						response.EslintPluginEntries[0].RuleNames = []string{"different"}
					}
				}
				return response, nil
			})
			cli.Start()
			peer.Start()
			loader := &ipcConfigModuleLoader{channel: cli, warmup: test.warmup}
			_, err := loader.ActivateConfigs(context.Background(), discovery.ConfigActivationRequest{
				ProtocolVersion: discovery.ConfigDiscoveryProtocolVersion,
				TransactionID:   "selected", EffectiveConfigIDs: []string{"root"},
			})
			if err != nil {
				t.Fatal(err)
			}
			wantKinds := []ipc.MessageKind{kindActivateConfigs}
			if test.warmup {
				wantKinds[0] = kindPrepareConfigs
			}
			if test.warmup && test.plugins {
				if loader.completeActivation == nil {
					t.Fatal("plugin preparation has no completion barrier")
				}
				if err := loader.completeActivation(); (err != nil) != (test.fail || test.mismatch) {
					t.Fatalf("complete activation: %v", err)
				}
				wantKinds = append(wantKinds, kindActivateConfigs)
			} else if loader.completeActivation != nil {
				t.Fatal("already activated config has a redundant barrier")
			}
			if !slices.Equal(kinds, wantKinds) {
				t.Fatalf("requests = %v, want %v", kinds, wantKinds)
			}
		})
	}
}
